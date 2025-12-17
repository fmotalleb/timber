FROM scratch
COPY timber /
ENTRYPOINT ["/timber"]
CMD ["--config=/config.toml"]
