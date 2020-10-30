FROM centos:7
ARG CONFIG=config.yml
RUN mkdir /app 
WORKDIR /app/

COPY main /app/log-formatter
COPY $CONFIG /app/config.yml
