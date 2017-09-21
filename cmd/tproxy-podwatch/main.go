package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// The annotation to look for
var annotation string

// The IP of the host this is running on.
var thisIP string

// The port number to redirect traffic to.
var hostPortString string

type Controller struct {
	indexer  cache.Indexer
	queue    workqueue.RateLimitingInterface
	informer cache.Controller
}

func NewController(queue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller) *Controller {
	return &Controller{
		informer: informer,
		indexer:  indexer,
		queue:    queue,
	}
}

func (c *Controller) processNextItem() bool {
	// Wait until there is a new item in the working queue
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two pods with the same key are never processed in
	// parallel.
	defer c.queue.Done(key)

	// Invoke the method containing the business logic
	err := c.syncFirewall(key.(string))
	// Handle the error if something went wrong during the execution of the business logic
	c.handleErr(err, key)
	return true
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *Controller) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.queue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if c.queue.NumRequeues(key) < 5 {
		glog.Infof("Error syncing pod %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	glog.Infof("Dropping pod %q out of the queue: %v", key, err)
}

func (c *Controller) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.queue.ShutDown()
	glog.Info("Starting Pod controller")

	go c.informer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	glog.Info("Stopping Pod controller")
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}

func (c *Controller) syncFirewall(key string) error {
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	podName := strings.Split(key, "/")[1]

	indices, err := checkFirewall(podName)
	if err != nil {
		log.Printf("Error checking firewall for existing rules on pod %s: %v\n", podName, err)
		return err
	}

	// If the entry doesn't exist then the pod has been deleted, remove any matching firewall rules.
	// Note that at this point we don't know what node the pod was running on so every node running
	// this controller will search for and remove matching firewall rules.
	if !exists {
		if len(indices) > 0 {
			log.Printf("Removing tproxy firewall rule for pod %s, chain index %v\n", podName, indices)

			// When removing rules by number, the table is reordered after each action.
			// Reverse list of indices so that they are removed in reverse order and the numbers match.
			for i, j := 0, len(indices)-1; i < j; i, j = i+1, j-1 {
				indices[i], indices[j] = indices[j], indices[i]
			}

			for i := range indices {
				if err := removeFirewall(indices[i]); err != nil {
					log.Printf("Error removing firewall rule number %d for pod %s: %v", indices[i], podName, err)
					return err
				}
			}
		}
		return nil
	}

	pod := obj.(*v1.Pod)
	podIP := pod.Status.PodIP

	// Only process pods running on the same host.
	if pod.Status.HostIP != thisIP {
		return nil
	}

	// Only process if podIP is known.
	if podIP == "" {
		return nil
	}

	// Only process pods with the annotation set.
	a := pod.ObjectMeta.GetAnnotations()
	_, ok := a[annotation]
	if !ok {
		log.Printf("Required annotation missing on pod %s; skipping firewall adjustment", podName)
		return nil
	}

	if len(indices) == 0 {
		log.Printf("Adding tproxy firewall rule for pod %s, %s\n", podName, podIP)

		// This comment is applied to the rule and used to find existing rules.
		comment := fmt.Sprintf("tproxy-%s", podName)

		if err := addFirewall(podIP, comment); err != nil {
			log.Printf("Error adding firewall rule for pod %s: %v\n", podName, err)
			return err
		}
	} else {
		if pod.DeletionTimestamp == nil {
			log.Printf("Firewall rule exists for pod %s, %s, skipping update.\n", podName, podIP)
		}
	}

	return nil
}

// Look for existing firewall entry given pattern string.
func checkFirewall(pattern string) ([]int, error) {

	// Using "-w" arg to prevent concurrency issues in iptables.
	cmd := exec.Command("iptables", "-w", "-t", "nat", "-L", "PREROUTING", "--line-numbers")
	output := &bytes.Buffer{}
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	// Capture the table entry number for lines containing the pattern.
	re, err := regexp.Compile(fmt.Sprintf(`([0-9]+).*REDIRECT.*%s.*`, pattern))
	if err != nil {
		log.Printf("Error compiling regexp")
		return nil, err
	}

	// Find all matches, create and return list of matching rule numbers.
	res := re.FindAllStringSubmatch(string(output.Bytes()), -1)
	if len(res) > 0 {
		var indices []int

		for j := range res {
			i, err := strconv.Atoi(res[j][1])
			if err != nil {
				return nil, err
			}
			indices = append(indices, i)
		}
		return indices, nil
	}
	return nil, nil
}

func addFirewall(ip, comment string) error {

	// Using "-w" arg to prevent concurrency issues in iptables.

	// Port 443
	cmd := exec.Command("iptables", "-w", "-t", "nat", "-A", "PREROUTING", "-s", ip, "-p", "tcp", "--dport", "443", "-j", "REDIRECT", "-m", "comment", "--comment", comment, "--to", hostPortString)
	if err := cmd.Run(); err != nil {
		return err
	}

	// Port 80
	cmd = exec.Command("iptables", "-w", "-t", "nat", "-A", "PREROUTING", "-s", ip, "-p", "tcp", "--dport", "80", "-j", "REDIRECT", "-m", "comment", "--comment", comment, "--to", hostPortString)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func removeFirewall(index int) error {

	// Using "-w" arg to prevent concurrency issues in iptables.

	// Note that this causes the table to reorder.
	num := strconv.Itoa(index)
	cmd := exec.Command("iptables", "-w", "-t", "nat", "-D", "PREROUTING", num)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func main() {
	var namespace string
	var hostPort int

	flag.StringVar(&namespace, "namespace", "default", "the namespace to monitor")
	flag.StringVar(&annotation, "annotation", "initializer.kubernetes.io/tproxy", "the pod annotation to match on")
	flag.IntVar(&hostPort, "hostPort", 8080, "The host port to redirect http/s traffic to.")
	flag.Parse()

	hostPortString = strconv.Itoa(hostPort)

	log.Println("Starting the Kubernetes pod watcher...")

	myIP, err := utilnet.ChooseHostInterface()
	if err != nil {
		log.Fatalf("Could not determine this ip: %v", err)
	}
	thisIP = myIP.String()

	log.Printf("Processing pods on host node: %s", thisIP)

	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		log.Fatal(err)
	}

	// create the pod watcher
	podListWatcher := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", namespace, fields.Everything())

	// create the workqueue
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the pod key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might see a newer version
	// of the Pod than the version which was responsible for triggering the update.
	indexer, informer := cache.NewIndexerInformer(podListWatcher, &v1.Pod{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
	}, cache.Indexers{})

	controller := NewController(queue, indexer, informer)

	// Now let's start the controller
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	// Wait forever
	select {}
}
