# fail2ban-exporter for prometheus

A prometheus exporter for fail2ban listerning on :8089, to get the metrics, the program will need to run as root, or have sudo permissions for fail2ban-client.

There are two routes 

`/metrics` is the default for prometheus and will return metrics about all known fail2ban jails

`/probe/{jail}` ie: `/probe/sshd` will return the metrics for just the named jail, in this example the sshd jail

## Example Output

### /metrics

```
# HELP fail2ban_current_banned Number of current banned IPs
# TYPE fail2ban_current_banned gauge
fail2ban_current_banned{jail="sshd"} 0
fail2ban_current_banned{jail="traefik-auth"} 0
# HELP fail2ban_current_failed Number of failed attempts
# TYPE fail2ban_current_failed gauge
fail2ban_current_failed{jail="sshd"} 0
fail2ban_current_failed{jail="traefik-auth"} 0
# HELP fail2ban_total_banned Total Number of banned IPs
# TYPE fail2ban_total_banned gauge
fail2ban_total_banned{jail="sshd"} 0
fail2ban_total_banned{jail="traefik-auth"} 0
# HELP fail2ban_total_failed Total Number of failed attempts
# TYPE fail2ban_total_failed gauge
fail2ban_total_failed{jail="sshd"} 0
fail2ban_total_failed{jail="traefik-auth"} 0
```

### /probe/sshd
```
# HELP fail2ban_current_banned Number of current banned IPs
# TYPE fail2ban_current_banned gauge
fail2ban_current_banned{jail="sshd"} 0
# HELP fail2ban_current_failed Number of failed attempts
# TYPE fail2ban_current_failed gauge
fail2ban_current_failed{jail="sshd"} 0
# HELP fail2ban_total_banned Total Number of banned IPs
# TYPE fail2ban_total_banned gauge
fail2ban_total_banned{jail="sshd"} 0
# HELP fail2ban_total_failed Total Number of failed attempts
# TYPE fail2ban_total_failed gauge
fail2ban_total_failed{jail="sshd"} 0
```
## Security Note

This code is currently work in progress and should not be considered production ready, the service is listerning on 0.0.0.0 and will need root permissions, use with care. 

## Other Notes

This work was inspired by https://github.com/jangrewe/prometheus-fail2ban-exporter but I wanted to have a http server to query rather than using prometheus's file reading.

I am not a programer and am still learning go, there are naturally optimisations to the code that I am happy to accept PRs for.