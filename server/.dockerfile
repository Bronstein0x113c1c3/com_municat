FROM ubuntu
RUN sysctl -w net.core.rmem_max=250000
RUN sysctl -w net.core.rmem_max=250000
COPY server /bin
WORKDIR /bin
EXPOSE 8080/udp
CMD [ "server" ]
