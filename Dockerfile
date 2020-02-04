FROM centos:8

LABEL author="xooooooox" description="file upload"

WORKDIR /ups

COPY . /ups

RUN ["/bin/bash", "-c", "./docker/run.bash"]

VOLUME ["/var/www/uploads", "/var/www/static"]

EXPOSE 80 8001

ENTRYPOINT ["ups"]

