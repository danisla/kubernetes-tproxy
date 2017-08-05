package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/healthz", healthzHandler)
	http.HandleFunc("/tproxy", tproxyHandler)

	s := &http.Server{
		Addr:           ":9000",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}

func tproxyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Only POST requests accepted.")
		return
	}

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Could not parse form data")
		return
	}
	action := r.FormValue("action")
	if action == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Missing form data: action (add/remove)")
		return
	}
	podIP := r.FormValue("pod_ip")
	if podIP == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Missing form data: pod_ip")
		return
	}
	podName := r.FormValue("pod_name")
	if podName == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Missing form data: pod_name")
		return
	}

	if action == "add" {
		log.Printf("Adding tproxy rule for pod %s, %s\n", podName, podIP)
		comment := fmt.Sprintf("tproxy-%s", podName)
		if err := addTProxy(podIP, comment, "-A"); err != nil {
			fmt.Printf("Error adding tproxy rule: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error adding tproxy rule for pod ip %s, %s\n", podIP, podName)
			return
		}
	} else if action == "remove" {
		log.Printf("Removing tproxy rule for pod %s, %s\n", podName, podIP)
		comment := fmt.Sprintf("tproxy-%s", podName)
		if err := addTProxy(podIP, comment, "-D"); err != nil {
			fmt.Printf("Error removing tproxy rule: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error removing tproxy rule for pod ip %s, %s\n", podIP, podName)
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Unsupported action: %s", action)
		return
	}

	fmt.Fprintln(w, "OK")
}

func addTProxy(ip, comment, action string) error {

	// Port 443
	// cmd := exec.Command("iptables", "-t", "nat", action, "PREROUTING", "-s", ip, "-p", "tcp", "--dport", "443", "-j", "REDIRECT", "-m", "comment", "--comment", comment, "--to", "8080")
	// if err := cmd.Run(); err != nil {
	// 	return err
	// }

	// Port 80
	// cmd = exec.Command("iptables", "-t", "nat", action, "PREROUTING", "-s", ip, "-p", "tcp", "--dport", "80", "-j", "REDIRECT", "-m", "comment", "--comment", comment, "--to", "8080")
	// if err := cmd.Run(); err != nil {
	// 	return err
	// }

	return nil
}
