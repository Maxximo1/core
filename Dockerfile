FROM ubuntu:16.04

COPY sonmworker /sonm/
COPY worker.yaml /sonm/
COPY keys /sonm/