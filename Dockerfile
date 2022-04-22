FROM moby/buildkit:v0.9.3
WORKDIR /dcp
COPY dcp README.md /dcp/
ENV PATH=/dcp:$PATH
ENTRYPOINT [ "/bhojpur/dcp" ]
