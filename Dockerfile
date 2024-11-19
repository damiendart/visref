# Copyright (C) Damien Dart, <damiendart@pobox.com>.
# This file is distributed under the MIT licence. For more information,
# please refer to the accompanying "LICENCE" file.

FROM golang:1.22-alpine AS build
RUN apk add build-base git nodejs upx go-task
WORKDIR /build
COPY . .
RUN go-task build-slim
RUN upx -9 visref

FROM alpine:latest
COPY --from=build /build/visref /usr/local/bin/visref
EXPOSE 4444
ENTRYPOINT ["/usr/local/bin/visref"]
