# Project-Description

This folder contains the **academic foundation** and **design documentation** for the Master's thesis project on blockchain-based Guarantee of Origin (GO) scheme interoperability for energy carrier conversion.

## Contents

### `README-files/`
Step-by-step setup guides for deploying the Hyperledger Fabric network:
- **Virtual-Machine-Setup.md** — Google Cloud VM provisioning (e2-standard-2, Ubuntu 22.04)
- **Installing-Prerequisites.md** — Docker, Go 1.22.1, Node.js 18, Fabric binaries
- **Bringing-up-the-network.md** — CA creation, crypto material generation, orderer/peer startup, channel creation
- **Deploying-and-commiting-the-Chaincode.md** — Chaincode lifecycle: package → install → approve → commit
- **Executing-network-commands.md** — Full transaction walkthrough (eGO creation, transfer, hydrogen conversion, cancellation)
- **Connect-Caliper.md** — Hyperledger Caliper performance benchmarking setup

### `Literature-Review/`
- **Conversion-Issuance-basicinfo.md** — Comprehensive review of GO schemes: legal framework, book-and-claim model, stakeholders, cross-border/cross-carrier interoperability, conversion rules (attribute inheritance, emission compounding, conversion efficiency), and proposed IT architectures
- **Requirements-Analysis.md** — Multi-vocal literature review identifying 8 key issues (low incentivization, lack of harmonization, high transaction costs, complexity, temporal/spatial decoupling, unreliable data, lack of transparency, privacy concerns) mapped to DLT-derived requirements
- **Table1.png / Table2.png** — Summary tables for the literature review

### `platform-agnostic-architecture/`
- **platform-agnostic-architecture.md** — High-level conceptual architecture designed before the Fabric implementation as part of the Design Science Research methodology
- **Slide1-9.png** — Architecture diagrams illustrating the platform-agnostic design
