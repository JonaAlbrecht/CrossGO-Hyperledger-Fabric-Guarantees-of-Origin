#!/bin/bash
# enroll-device-identity.sh — Register and enroll device identities for Caliper benchmarking
# ADR-027: Creates device identities with X.509 attributes required by CreateElectricityGO/CreateBiogasGO
# Prerequisites: Fabric CA client binary, CAs running (docker-compose-ca.yaml)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CALIPER_DIR="$(dirname "$SCRIPT_DIR")"
REPO_DIR="$(cd "$CALIPER_DIR/../.." && pwd)"
NETWORK_DIR="$REPO_DIR/network"
DEVICE_CERTS_DIR="$CALIPER_DIR/device-identities"

export PATH="${REPO_DIR}/fabric-bin/bin:$PATH"

RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m'

log()  { echo -e "${GREEN}[✓]${NC} $1"; }
info() { echo -e "${CYAN}[→]${NC} $1"; }

# ─── Clean previous device certs ──────────────────────────────────────────
rm -rf "$DEVICE_CERTS_DIR"
mkdir -p "$DEVICE_CERTS_DIR"

# ─── Check if CAs are running ────────────────────────────────────────────
info "Checking if Fabric CAs are running..."
if ! docker ps --format '{{.Names}}' | grep -q 'ca.eproducer1'; then
    info "Starting Fabric CAs..."
    docker compose -f "$NETWORK_DIR/docker/docker-compose-ca.yaml" up -d
    sleep 5
fi
log "Fabric CAs running"

# ─── Helper: Get CA TLS cert path ────────────────────────────────────────
get_ca_tls_cert() {
    local org=$1
    echo "$NETWORK_DIR/organizations/fabric-ca/${org}/tls-cert.pem"
}

# ─── Enroll eproducer1 CA admin ──────────────────────────────────────────
info "Enrolling eproducer1 CA admin..."
export FABRIC_CA_CLIENT_HOME="$DEVICE_CERTS_DIR/eproducer1-ca-admin"
mkdir -p "$FABRIC_CA_CLIENT_HOME"

fabric-ca-client enroll -u https://admin:adminpw@localhost:8054 \
    --caname ca-eproducer1 \
    --tls.certfiles "$(get_ca_tls_cert eproducer1)" \
    2>&1 || true
log "eproducer1 CA admin enrolled"

# ─── Register electricity device identity ─────────────────────────────────
info "Registering electricity SmartMeter device identity..."
fabric-ca-client register \
    --caname ca-eproducer1 \
    --id.name device-electricity-meter \
    --id.secret devicepw \
    --id.type client \
    --id.attrs 'electricitytrustedDevice=true:ecert,maxEfficiency=100:ecert,emissionIntensity=50:ecert,technologyType=solar:ecert,hf.Registrar.Roles=client' \
    --tls.certfiles "$(get_ca_tls_cert eproducer1)" \
    2>&1 || true
log "Electricity device identity registered"

# ─── Register biogas device identity ──────────────────────────────────────
info "Registering biogas SmartMeter device identity..."
fabric-ca-client register \
    --caname ca-eproducer1 \
    --id.name device-biogas-meter \
    --id.secret devicepw \
    --id.type client \
    --id.attrs 'biogastrustedDevice=true:ecert,maxOutput=500:ecert,technologyType=anaerobic_digestion:ecert,hf.Registrar.Roles=client' \
    --tls.certfiles "$(get_ca_tls_cert eproducer1)" \
    2>&1 || true
log "Biogas device identity registered"

# ─── Enroll electricity device identity ───────────────────────────────────
info "Enrolling electricity device identity..."
export FABRIC_CA_CLIENT_HOME="$DEVICE_CERTS_DIR/electricity-device"
mkdir -p "$FABRIC_CA_CLIENT_HOME"

fabric-ca-client enroll -u https://device-electricity-meter:devicepw@localhost:8054 \
    --caname ca-eproducer1 \
    --tls.certfiles "$(get_ca_tls_cert eproducer1)" \
    --enrollment.attrs "electricitytrustedDevice,maxEfficiency,emissionIntensity,technologyType" \
    --mspdir "$DEVICE_CERTS_DIR/electricity-device/msp" \
    2>&1
log "Electricity device identity enrolled"

# ─── Enroll biogas device identity ────────────────────────────────────────
info "Enrolling biogas device identity..."
export FABRIC_CA_CLIENT_HOME="$DEVICE_CERTS_DIR/biogas-device"
mkdir -p "$FABRIC_CA_CLIENT_HOME"

fabric-ca-client enroll -u https://device-biogas-meter:devicepw@localhost:8054 \
    --caname ca-eproducer1 \
    --tls.certfiles "$(get_ca_tls_cert eproducer1)" \
    --enrollment.attrs "biogastrustedDevice,maxOutput,technologyType" \
    --mspdir "$DEVICE_CERTS_DIR/biogas-device/msp" \
    2>&1
log "Biogas device identity enrolled"

# ─── Fix private key filenames ────────────────────────────────────────────
# The fabric-ca-client produces keys with long random names; Caliper needs a stable reference.
for identity in electricity-device biogas-device; do
    KEYSTORE="$DEVICE_CERTS_DIR/$identity/msp/keystore"
    if [ -d "$KEYSTORE" ]; then
        KEY_FILE=$(ls "$KEYSTORE/"*_sk 2>/dev/null | head -n 1)
        if [ -n "$KEY_FILE" ]; then
            cp "$KEY_FILE" "$KEYSTORE/priv_sk"
            log "Fixed key for $identity"
        fi
    fi
done

echo ""
echo "=============================================="
echo "  Device Identities Ready for Caliper"
echo "=============================================="
echo ""
echo "Electricity device cert: $DEVICE_CERTS_DIR/electricity-device/msp/signcerts/cert.pem"
echo "Electricity device key:  $DEVICE_CERTS_DIR/electricity-device/msp/keystore/priv_sk"
echo "Biogas device cert:      $DEVICE_CERTS_DIR/biogas-device/msp/signcerts/cert.pem"
echo "Biogas device key:       $DEVICE_CERTS_DIR/biogas-device/msp/keystore/priv_sk"
echo ""
echo "These identities have the following X.509 attributes:"
echo "  electricity: electricitytrustedDevice=true, maxEfficiency=100, emissionIntensity=50, technologyType=solar"
echo "  biogas: biogastrustedDevice=true, maxOutput=500, technologyType=anaerobic_digestion"
