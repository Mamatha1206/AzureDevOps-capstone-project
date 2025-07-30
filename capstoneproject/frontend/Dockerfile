# Build stage
FROM node:18-alpine AS build

WORKDIR /app

COPY package*.json ./
RUN npm install

COPY . .

# Production stage using distroless
FROM gcr.io/distroless/nodejs:18

WORKDIR /app

# Copy only necessary files from build stage
COPY --from=build /app/index.js ./
COPY --from=build /app/server.js ./
COPY --from=build /app/node_modules ./node_modules
COPY --from=build /app/public ./public

EXPOSE 3000

# The distroless image directly runs the JavaScript file.
CMD ["index.js"]
