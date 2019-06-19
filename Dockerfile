FROM alpine:3.6

LABEL Author="Denis Sheshnev <denis.sheshnev@lamoda.ru>"

COPY gonkey /bin/gonkey
ENTRYPOINT ["/bin/gonkey"]
CMD ["-spec=/gonkey/swagger.yaml", "-host=${HOST_ARG}"]
