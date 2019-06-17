FROM alpine:3.6

LABEL Author="Denis Sheshnev <denis.sheshnev@lamoda.ru>"

COPY output /bin/output
ENTRYPOINT ["/bin/output"]
CMD ["-spec=/gonkey/swagger.yaml", "-host=${HOST_ARG}"]
