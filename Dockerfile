FROM scratch
RUN apt-get update
COPY build/discovery /usr/local/bin/discovery
CMD /usr/local/bin/discovery