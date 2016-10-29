FROM golang
ENV DEBUG false
ENV PORT 8080
ENV USEOAUTH false
ENV GITHUB_CLIENT_ID ""
ENV GITHUB_CLIENT_SECRET ""
EXPOSE 8080
ADD . /go
ENTRYPOINT echo "{\"debug\":${DEBUG},\"port\":${PORT},\"useOAuth\":${USEOAUTH},\"github\":{\"client_id\":\"${GITHUB_CLIENT_ID}\",\"client_secret\":\"${GITHUB_CLIENT_SECRET}\"}}" > config.json && ./install.sh && ./gobang

