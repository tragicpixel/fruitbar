FROM node:16.10.0-alpine AS dev

WORKDIR /app

ENV PATH /app/node_modules/.bin:$PATH

# install app dependencies
COPY fruitbar-ui/package.json ./
COPY fruitbar-ui/package-lock.json ./
RUN npm install --silent
RUN npm install react-scripts@3.4.1 -g --silent
# Note: --silent can cause you to miss errors, remove it from the npm installs in case of troubleshooting

ADD fruitbar-ui/ ./

# this should definitely be used in dev so I can test on my local machine
EXPOSE 3000

# start app
CMD ["npm", "start"]