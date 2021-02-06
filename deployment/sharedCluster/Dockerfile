FROM mongo:4.4.3

COPY setupSharedCluster.sh setupSharedCluster.sh

RUN chmod +x setupSharedCluster.sh

ENTRYPOINT ["./setupSharedCluster.sh"]