IMAGE_NAME="forum_image"
CONTAINER_NAME="forum_container"
PORT="10443"

echo "Building Docker image"
docker image build -t $IMAGE_NAME .
echo "Running Docker container"
docker run -dp $PORT:$PORT --name $CONTAINER_NAME $IMAGE_NAME
echo "Container running on port $PORT"
echo "https://localhost:$PORT/"

