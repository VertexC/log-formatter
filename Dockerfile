FROM centos:7
RUN mkdir /app 
WORKDIR /app/

COPY main /app/log-formatter
