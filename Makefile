images: image-sidecar image-initializer image-podwatch

image-sidecar:
	cd sidecar && \
	gcloud container builds submit --config cloudbuild.yaml .

image-initializer:
	gcloud container builds submit --config cloudbuild-initializer.yaml .

image-podwatch:
	gcloud container builds submit --config cloudbuild-podwatch.yaml .