FROM alpine:3.6 as brotli
RUN apk --no-cache add bash build-base cmake musl-dev git
RUN git clone -b v0.6.0 https://github.com/google/brotli.git /tmp/brotli
RUN cd /tmp/brotli && ./configure-cmake && make && make install


FROM alpine:3.6 as base
COPY --from=brotli /usr/local/lib/libbrotlienc.so.0.6.0 /usr/local/lib/libbrotlienc.so.0.6.0
COPY --from=brotli /usr/local/lib/libbrotlienc.so /usr/local/lib/libbrotlienc.so
COPY --from=brotli /usr/local/lib/libbrotlidec.so.0.6.0 /usr/local/lib/libbrotlidec.so.0.6.0
COPY --from=brotli /usr/local/lib/libbrotlidec.so /usr/local/lib/libbrotlidec.so
COPY --from=brotli /usr/local/lib/libbrotlicommon.so.0.6.0 /usr/local/lib/libbrotlicommon.so.0.6.0
COPY --from=brotli /usr/local/lib/libbrotlicommon.so /usr/local/lib/libbrotlicommon.so
COPY --from=brotli /usr/local/include/brotli/encode.h /usr/local/include/brotli/encode.h
COPY --from=brotli /usr/local/include/brotli/types.h /usr/local/include/brotli/types.h
COPY --from=brotli /usr/local/include/brotli/port.h /usr/local/include/brotli/port.h
COPY --from=brotli /usr/local/include/brotli/decode.h /usr/local/include/brotli/decode.h
COPY --from=brotli /usr/local/lib/pkgconfig/libbrotlicommon.pc /usr/local/lib/pkgconfig/libbrotlicommon.pc
COPY --from=brotli /usr/local/lib/pkgconfig/libbrotlidec.pc /usr/local/lib/pkgconfig/libbrotlidec.pc
COPY --from=brotli /usr/local/lib/pkgconfig/libbrotlienc.pc /usr/local/lib/pkgconfig/libbrotlienc.pc


FROM base as builder
RUN apk --no-cache add build-base go musl-dev git
WORKDIR /root/src/github.com/badgerodon/grpcsimulator
ENV GOPATH /root
RUN git clone -b caleb/net https://github.com/badgerodon/gopherjs.git /root/src/github.com/gopherjs/gopherjs
RUN go get -v github.com/gopherjs/gopherjs

COPY build.go .
COPY main.go .
COPY builder/ ./builder
COPY vendor/ ./vendor
RUN go build -o /root/bin/app .


FROM base as runner
RUN apk --no-cache add ca-certificates
WORKDIR /root
ENV GOPATH /root
ENV GOOGLE_APPLICATION_CREDENTIALS /root/gcloud.credentials
ENV PORT=80
COPY --from=builder /root/bin/app /root/bin/app
COPY --from=builder /root/bin/gopherjs /root/bin/gopherjs
COPY --from=builder /root/src/github.com/gopherjs/gopherjs /root/src/github.com/gopherjs/gopherjs
COPY --from=builder /root/src/github.com/badgerodon/grpcsimulator /root/src/github.com/badgerodon/grpcsimulator
CMD ["/root/bin/app"]


EXPOSE 80
