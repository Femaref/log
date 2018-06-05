#! /bin/sh

start-stop-daemon -q -m --oknodo --start --exec /main --pidfile /pid --background
sleep 1
start-stop-daemon -q --stop --pidfile /pid
sleep 1
echo trace
cat trace 
echo ---
echo log
cat foo.log
echo ---
echo stderr
cat stderr-panic.log