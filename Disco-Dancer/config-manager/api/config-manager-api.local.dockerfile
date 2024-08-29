# FROM geiconp.azurecr.io/billing/alpine-cert:1.0
FROM artifactory-pd-infra.aks.aze1.cloud.geico.net/mvp-billing-container-all/billing/base/alpine:3.19
COPY --chown=appuser:appuser api/publish/ /app
RUN chmod -R 0500 /app
EXPOSE 30000
WORKDIR /app
ENTRYPOINT ["/app/config-manager-api"]