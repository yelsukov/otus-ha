### STAGE 1: Build ###
FROM node:14.15-alpine AS builder
WORKDIR /usr/src/app
COPY package.json package-lock.json ./
RUN npm install

ENV PATH="./node_modules/.bin:$PATH"

COPY . ./
RUN ng build --prod

### STAGE 2: Run ###
FROM nginx:1.19.6-alpine
COPY deploy/nginx.conf /etc/nginx/nginx.conf
COPY --from=builder /usr/src/app/dist/frontend /usr/share/nginx/html
