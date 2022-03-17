#!/bin/bash
REMOTE_HOST=127.0.0.1

XDEBUG_CONF=$(php -i | grep xdebug.ini | cut -d ',' -f1)
echo zend_extension=xdebug.so > ${XDEBUG_CONF}
echo xdebug.client_port=${1:-9003} >> ${XDEBUG_CONF}
echo xdebug.client_host=${REMOTE_HOST} >> ${XDEBUG_CONF}
echo xdebug.idekey=${2:-PHPSTORM} >> ${XDEBUG_CONF}
echo xdebug.mode=debug >> ${XDEBUG_CONF}

apache2ctl -k restart

echo ""
echo "You'll probably have to change 'xdebug.client_host' if you are running on macOS or Windows. Don't forget to restart apache2 in case of any changes: apache2ctl -k restart."
echo "Your xdebug config ($XDEBUG_CONF):"
echo ""
cat ${XDEBUG_CONF}
echo ""