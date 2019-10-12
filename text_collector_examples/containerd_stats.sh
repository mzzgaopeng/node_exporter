#!/bin/bash

# Check that containerd is running
pidof docker-containerd > /dev/null 2>&1
RC=$?
RS=1
if [ $RC -ne 0 ]
then
  RS=0
fi

echo "# HELP get containerd_status from shell"
echo "# TYPE containerd_status gauge"
echo "shell_containerd_status $RS"