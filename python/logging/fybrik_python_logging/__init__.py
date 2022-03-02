#
# Copyright 2022 IBM Corp.
# SPDX-License-Identifier: Apache-2.0
#
import logging
import json_log_formatter
import time

# FybrikFormatter constants
FybrikAppUUID = 'app.fybrik.io/app-uuid'
Level         = 'level'
Message       = 'message'
Time          = 'time'
Caller        = 'caller'
FuncName      = 'funcName'

Error         = 'error'
DataSetID     = 'DataSetID'
ForUser       = 'ForUser'

TRACE = 5

logging.TRACE = TRACE
logging.addLevelName(TRACE, "TRACE")

def trace(self, msg, *args, **kwargs):
    self._log(TRACE, msg, args, **kwargs)

logging.Logger.trace = trace

logger = logging.getLogger('')
app_uuid = ''

class FybrikFormatter(json_log_formatter.JSONFormatter):
    def json_record(self, message: str, extra: dict, record: logging.LogRecord) -> dict:
        extra[Message] = message
        extra[Level] = record.levelname
        extra[Caller] = record.filename + ':' + str(record.lineno)
        extra[FuncName] = record.funcName
        extra[Time] = time.strftime('%Y-%m-%dT%X%z', time.localtime(record.created))
        extra[FybrikAppUUID] = app_uuid
        return extra

def init_logger(loglevel_arg, app_uuid_str, module_name):
    global app_uuid
    app_uuid = app_uuid_str
    loglevel = getattr(logging, loglevel_arg, logging.WARNING)
    logger.name = module_name
    logger.setLevel(loglevel)
    ch = logging.StreamHandler()
    ch.setLevel(loglevel)
    ch.setFormatter(FybrikFormatter())
    logger.addHandler(ch)
