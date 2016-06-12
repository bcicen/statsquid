FROM alpine:latest

RUN apk add --no-cache python ca-certificates wget && \
    wget "https://bootstrap.pypa.io/get-pip.py" -O - | python

COPY requirements.txt /
RUN pip install -r requirements.txt

COPY . /usr/src/app/
RUN cd /usr/src/app/ && \
    python setup.py install

ENTRYPOINT [ "statsquid" ]
