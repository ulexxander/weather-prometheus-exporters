
DOCKER_REPO="ulexxander/open-weather-prometheus-exporter"
GIT_TAG=$(git describe --tags)
IMAGE_TAG="$DOCKER_REPO:$GIT_TAG"
LATEST_TAG="$DOCKER_REPO:latest"

build() {
  docker build -t $IMAGE_TAG .
  docker build -t $LATEST_TAG .
}

push() {
  docker push $IMAGE_TAG
  docker push $LATEST_TAG
}
