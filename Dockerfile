FROM python:2

COPY requirements.txt /
RUN pip install -r requirements.txt

ADD . /usr/src/app/
RUN cd /usr/src/app/ && \
    python setup.py install

ENTRYPOINT [ "statsquid" ]
