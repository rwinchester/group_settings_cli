function sanitize_pem {
  awk '{printf "%s\\n",$0} END {print ""}' $1
}

export GSCLI_CLIENT_CERT='"'$(sanitize_pem /Users/rob/certs/02020014-cert.pem)'"'
export GSCLI_CLIENT_KEY='"'$(sanitize_pem /Users/rob/certs/02020014-key.pem)'"'
export GSCLI_BACKEND_URI='https://lwdev2.lightrules.net:443'
export GSCLI_ROUTER_URI='wss://lwdev2.lightrules.net:8444'
