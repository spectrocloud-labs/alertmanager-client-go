#!/bin/bash

# Script to generate self-signed certificates for testing TLS with Alertmanager

set -e

CERT_DIR="./certs"
mkdir -p "$CERT_DIR"

echo "Generating self-signed certificates for testing..."

# Generate CA private key
openssl genrsa -out "$CERT_DIR/ca-key.pem" 2048

# Generate CA certificate
openssl req -new -x509 -days 365 -key "$CERT_DIR/ca-key.pem" \
    -out "$CERT_DIR/ca.pem" \
    -subj "/C=US/ST=Test/L=Test/O=Alertmanager Test/CN=Test CA"

# Generate server private key
openssl genrsa -out "$CERT_DIR/server-key.pem" 2048

# Generate server certificate signing request
openssl req -new -key "$CERT_DIR/server-key.pem" \
    -out "$CERT_DIR/server.csr" \
    -subj "/C=US/ST=Test/L=Test/O=Alertmanager/CN=localhost"

# Create extensions file for SAN
cat > "$CERT_DIR/server-ext.cnf" <<EOF
subjectAltName = DNS:localhost,IP:127.0.0.1
EOF

# Sign the server certificate with CA
openssl x509 -req -days 365 \
    -in "$CERT_DIR/server.csr" \
    -CA "$CERT_DIR/ca.pem" \
    -CAkey "$CERT_DIR/ca-key.pem" \
    -CAcreateserial \
    -out "$CERT_DIR/server.pem" \
    -extfile "$CERT_DIR/server-ext.cnf"

# Clean up CSR and extensions file
rm "$CERT_DIR/server.csr" "$CERT_DIR/server-ext.cnf"

echo "✓ Certificates generated successfully in $CERT_DIR/"
echo "  • ca.pem - CA certificate (use with WithCustomCA)"
echo "  • server.pem - Server certificate"
echo "  • server-key.pem - Server private key"
