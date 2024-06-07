#!/bin/bash

LOG_DIR="output"

# Merge & sort logs
cat ${LOG_DIR:?}/Log*.log > ${LOG_DIR:?}/completeLog.log
cat ${LOG_DIR:?}/GoVector/LogFile* >> ${LOG_DIR:?}/GoVector/temp_log.log
echo -e "(?<timestamp>\\d*) (?<host>\\w*) (?<clock>.*)\\\\n(?<event>.*)\\n" > ${LOG_DIR:?}/GoVector/completeGoVectorLog.log
sed -n '/^.*}$/ {N; s/\n/$\^$/g; p;}' ${LOG_DIR:?}/GoVector/temp_log.log | sort | sed -n 's/\$\^\$/\n/g; p;' >> ${LOG_DIR:?}/GoVector/completeGoVectorLog.log
rm ${LOG_DIR:?}/GoVector/temp_log.log
