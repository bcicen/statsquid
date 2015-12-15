FROM alpine:3.2

RUN apk add --update python ca-certificates wget && \
    wget "https://bootstrap.pypa.io/get-pip.py" -O - | python && \
    rm /var/cache/apk/*

COPY requirements.txt /
RUN pip install -r requirements.txt

COPY . /usr/src/app/
RUN cd /usr/src/app/ && \
    python setup.py install

ENTRYPOINT [ "statsquid" ]
