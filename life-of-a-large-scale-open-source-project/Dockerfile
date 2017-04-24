FROM node:alpine

RUN apk add --no-cache \
	git

RUN npm install -g \
	bower \
	gulp

COPY . /usr/src/app
WORKDIR /usr/src/app

RUN npm install
RUN bower install --allow-root
