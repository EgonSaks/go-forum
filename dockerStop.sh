IMAGE_NAME="forum_image"
CONTAINER_NAME="forum_container"

echo "Stopping and removing Docker container"
docker stop $CONTAINER_NAME && docker rm $CONTAINER_NAME
echo "Deleting Docker image"
docker rmi $IMAGE_NAME
echo "Container and image deleted"