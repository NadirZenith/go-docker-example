# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/NadirZenith/gowiki

## Copy static files
#RUN cp /go/src/github.com/NadirZenith/gowiki/edit.html /go/bin
#RUN cp /go/src/github.com/NadirZenith/gowiki/view.html /go/bin
WORKDIR /go/src/github.com/NadirZenith/gowiki

# Build the outyet command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
#RUN go install github.com/NadirZenith/gowiki
RUN go install

# Run the outyet command by default when the container starts.
ENTRYPOINT /go/bin/gowiki

# Document that the service listens on port 8080.
EXPOSE 8080