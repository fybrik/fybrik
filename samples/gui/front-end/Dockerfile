FROM node:argon

RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app

COPY deployment/package.json /usr/src/app/

RUN npm install

COPY build /usr/src/app/build
COPY deployment/server.js /usr/src/app/

EXPOSE 3000

CMD [ "npm", "start" ]
