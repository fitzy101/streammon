# Streammon
A program that monitors streams of text (eg stdin, local file) and looks for lines matching a user defined regex.
When a match is found, a user-defined (shell) command is executed. The user can reference fields from the matched line
within the command.

Current latest version: v0.1.0

# Building streammon
Use the `go build` command:

```
$ go build -o streammon main.go
```
You can then add the executable to somewhere in your $PATH.
Streammon was written with go 1.8, but is test to work with versions above 1.6.

# Usage
```
$ streammon -h
Usage of streammon:
  -a string
        a quoted string of arguments to the command.
  -args string
        a quoted string of arguments to the command.
  -c string
        a command to run after a match is found.
  -command string
        a command to run after a match is found.
  -d string
        a delimeter to split a matchin line. (default " ")
  -delimeter string
        a delimeter to split a matchin line. (default " ")
  -f string
        a full path to a file to monitor.
  -file string
        a full path to a file to monitor.
  -r string
        a regular expression to match. (default ".*")
  -regexp string
        a regular expression to match. (default ".*")
```

## Arguments
The arguments provided to the command to be run when a match is found can reference the fields within the command via the token #{n}. Where n is the field number when split by the delimeter provided by -d. If #{0} is provided or the field doesn't exist, the entire line matched will be passed as the command's first argument.

## Example
As an example, we'll monitor a text stream via stdin, and run a command that writes matched lines to a file.
The stream will be the output from a isc-dhcp server, and we'll look for DHCPDISCOVER requests.

The output from journalctl is formatted as so:
```
Aug 5 10:37:19 fitzy dhcpd[21613]: DHCPREQUEST for 192.168.127.3 from 61:7c:db:fb:45:5e via br1
Aug 5 10:37:19 fitzy dhcpd[21613]: DHCPACK on 192.168.127.3 to 61:7c:db:fb:45:5e via br1
Aug 5 10:37:45 fitzy dhcpd[21613]: DHCPREQUEST for 192.168.127.11 from a0:d0:33:a9:7b:49 via br1
Aug 5 10:37:45 fitzy dhcpd[21613]: DHCPACK on 192.168.127.11 to a0:d0:33:a9:7b:49 via br1
```

The arguments for streammon:
```
$ journalctl -fu isc-dhcp-server | streammon -a "ip:#{8} mac:#{10}" -c ~/save-to-desktop.sh -r DHCPREQUEST
```

We're going to monitor the output from the isc-dhcp logs for any DHCPREQUEST that comes through, and pass the 8th field (the IP address) and 10th field (the MAC address) to the command. In this case, the command is a script that simply pushes the argument to a file.

```
$ cat ~/save-to-desktop.sh
#!/bin/bash
for var in "$@"
do
        printf "%s %s\n" $1 $2 >> /tmp/streammon.log
done
exit 0
```

The file `/tmp/streammon.log` will contain the IP address after the command runs.

```
$ cat /tmp/streammon.log
ip:10.1.100.3 mac:d1:87:f8:5f:9d:1f
ip:10.1.100.11 mac:d1:87:f8:5f:9e:1f
ip:10.1.100.6 mac:d1:87:f8:5f:9f:1f
ip:10.1.100.9 mac:d1:87:f8:5f:9d:9f
ip:10.1.100.8 mac:d1:87:f8:5e:9d:1f
```

## TODO
- Implement config file specification, including ability to monitor several streams at once.
- Provide flag for displaying output of monitored streams.
- Implement a timeout feature, where the command will be run after waiting for the specified timeout.

## Known issues
- Multiple flags will sometimes not be output correctly when they are not seperated by a space in the command line argument - ie they'll be unparsed in the output. They work as expected when they are seperated by a space.

# License
Streammon is licensed under the MIT license.
