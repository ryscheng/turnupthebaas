# Dockerfile for talek
FROM talek-base:latest
MAINTAINER Raymond Cheng <me@raymondcheng.net>
USER root

# Install OpenCL
RUN apt-get update && apt-get install -y opencl-headers ocl-icd-opencl-dev
WORKDIR /talek/pird/
RUN rm pird
RUN make

WORKDIR /talek
