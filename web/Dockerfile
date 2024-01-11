# Base Image
# checkov:skip=CKV_DOCKER_2: curl/wget is not part of the distroless image, because it is distroless
# checkov:skip=CKV_DOCKER_3: not needed for  distroless image, because it is distroless
# checkov:skip=CKV_DOCKER_7: there are no immutable tags for distroless images
FROM node:21.5.0-alpine3.19 as builder

# Set working directory
WORKDIR /app

# Install dependencies
COPY package*.json ./
RUN npm install --omit=dev

# Copy source code
COPY . .

# Build the application
RUN npm run build

# Final Image
FROM gcr.io/distroless/nodejs20-debian12:latest@sha256:c63fbd00d837baee8160c764124f9f33a9bea2b38e0c4d6306453e43ac6bf443

# Set working directory
WORKDIR /app

# Copy built files from builder stage
COPY --from=builder /app/build ./build
COPY --from=builder /app/node_modules ./node_modules
COPY server.js ./

# Expose port and start application
EXPOSE 80

CMD ["server.js"]
