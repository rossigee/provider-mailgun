FROM gcr.io/distroless/static:nonroot

# OpenContainers Image Spec Labels
LABEL org.opencontainers.image.title="provider-mailgun"
LABEL org.opencontainers.image.description="Crossplane provider for Mailgun email service management"
LABEL org.opencontainers.image.vendor="rossigee"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.source="https://github.com/rossigee/provider-mailgun"
LABEL org.opencontainers.image.url="https://github.com/rossigee/provider-mailgun"
LABEL org.opencontainers.image.documentation="https://github.com/rossigee/provider-mailgun/blob/master/README.md"

ARG TARGETOS
ARG TARGETARCH

COPY --chmod=0755 bin/${TARGETOS}_${TARGETARCH}/provider /usr/local/bin/provider
COPY --chmod=0644 package.yaml /package.yaml
COPY --chmod=0644 crds /crds

USER 65532:65532
ENTRYPOINT ["/usr/local/bin/provider"]
