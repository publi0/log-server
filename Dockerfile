FROM alpine:latest

WORKDIR /root/

COPY log-server .

CMD ["./log-server"]


# Build your Docker image
docker build -t log-server .

# Tag your Docker image
docker tag log-server:latest felipecanton/log-server:latest

# Log in to Docker Hub
docker login --username=yourusername --password=yourpassword

# Push your Docker image to Docker Hub
docker push felipecanton/log-server:latest