FROM docker.io/fyneio/fyne-cross-images:windows
USER root
RUN apt-get update && apt-get install -y \
    mingw-w64 \
    gcc \
    g++ \
  && apt-get clean && rm -rf /var/lib/apt/lists/*
USER 1000
