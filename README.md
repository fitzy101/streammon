# Streammon
A program that monitors streams of text (eg stdin, local file) and looks for lines matching a user defined regex.
When a match is found, a user-defined (shell) command is executed. The user can reference fields from the matched line
within the command.

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
Aug 5 10:37:19 fitzy dhcpd[21613]: DHCPREQUEST for 192.168.127.3 from 00:13:c6:00:71:5e via br0
Aug 5 10:37:19 fitzy dhcpd[21613]: DHCPACK on 192.168.127.3 to 00:13:c6:00:71:5e via br0
Aug 5 10:37:45 fitzy dhcpd[21613]: DHCPREQUEST for 192.168.127.11 from 52:54:00:98:9d:53 via br0
Aug 5 10:37:45 fitzy dhcpd[21613]: DHCPACK on 192.168.127.11 to 52:54:00:98:9d:53 via br0
```

The arguments for streammon:
```
journalctl -fu isc-dhcp-server | streammon -a "#{8}" -c ~/save-to-desktop.sh -r DHCPREQUEST
```

We're going to monitor the output from the isc-dhcp logs for any DHCPREQUEST that comes through, and pass the 8th field (the IP address) to the command. In this case, the command is a script that simply pushes the argument to a file.

```
$ cat ~/save-to-desktop.sh
#!/bin/bash
echo $1 >> ~/Desktop/mylog
exit 0
```

The file `mylog` will contain the IP address after the command runs.

```
$ cat mylog
192.168.100.8
192.168.100.17
192.168.100.6
192.168.100.5
192.168.100.9
192.168.100.8
192.168.100.7
192.168.100.3
192.168.100.11
192.168.100.17
192.168.100.5
192.168.100.6
```

## TODO
- Implement config file specification, including ability to monitor several streams at once.
- Provide flag for displaying output of monitored streams.
- Implement a timeout feature, where the command will be run after waiting for the specified timeout.