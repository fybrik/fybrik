# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

FROM python:3.6-slim

WORKDIR /usr/app

COPY requirements.txt .
RUN pip3 install --no-cache-dir -r requirements.txt

COPY . .

WORKDIR /usr/app/src

CMD [ "python3", "server.py" ]