# Node app with cache-busting copy order and no cache mount.
FROM node:latest
WORKDIR /app
COPY . .
RUN npm install
RUN apt-get update && apt-get install -y curl
CMD ["node", "server.js"]
