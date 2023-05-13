FROM alpine:latest
COPY out/cnetmon /usr/local/bin/cnetmon
RUN addgroup --gid 1000 cnetmon
RUN adduser --disabled-password --gecos "" --home "$(pwd)" --ingroup "cnetmon" --no-create-home --uid "1000" "cnetmon"
USER 1000:1000
CMD ["/usr/local/bin/cnetmon"]