FROM scratch
COPY kubectl-decode /usr/bin/kubectl-decode
ENTRYPOINT ["/usr/bin/kubectl-decode"]
