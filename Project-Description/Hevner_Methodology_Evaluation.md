# Literature Analysis — Recent Research on DLT-Based Energy Certificate Systems (2022–2025)

**Context:** ICIS 2026 paper "All Carriers Allowed: Design Principles for Distributed Ledger-Based Cross-Domain Data Sharing Between Guarantee of Origin Schemes"

**Method:** Systematic search via OpenAlex API — 30 queries across 8 thematic clusters, 857 unique results, filtered to 23 highly relevant papers since 2022.

**Date:** April 2026

---

## 1. Introduction

This analysis examines 23 recent publications (2022–2025) identified through a systematic API-based literature search as relevant to the design of DLT-based cross-carrier Guarantee of Origin (GO) scheme interoperability. The papers are organized into five thematic streams that correspond to the core dimensions of the ICIS 2026 paper: (1) GO scheme design and hydrogen certification, (2) blockchain-based renewable energy certificate systems, (3) Hyperledger Fabric performance and privacy, (4) double counting and residual mix challenges, and (5) broader blockchain-for-energy context. For each stream, we synthesize findings, identify convergences, and discuss implications for the paper's challenges (C1–C8), meta-requirements (MR1–MR4), and design principles (DP1–DP4).

---

## 2. Stream A — GO Scheme Design and Hydrogen Certification

This stream addresses the most directly relevant body of recent work: the design of certification and traceability systems for green hydrogen, biogas, and multi-carrier energy.

### Schmid, Ubacht & van Engelenburg (2024)

Schmid et al. present a DSR-based blockchain certification system for the EU hydrogen market in *Frontiers in Blockchain*. The paper directly addresses the same problem domain — certification of green hydrogen using DLT — and follows a comparable DSR methodology. The authors design a blockchain-based system to enable trustworthy and secure information sharing between untrusted parties for verifying hydrogen sustainability claims. This work provides the closest external validation of the ICIS paper's premise: that DLT is suited for certification systems where multiple parties must verify provenance without a single trusted authority. However, Schmid et al. focus on hydrogen as a single carrier rather than the cross-carrier conversion issuance problem, making the ICIS paper's contribution (cross-domain sharing *between* scheme registries) a clear extension.

**Relevance:** Validates the DSR approach to DLT-based certification; supports MR2 (Harmonization) and the need for DP5-like interoperability. Confirms market need for blockchain-based hydrogen certification.

### Heeß, Rockstuhl & Körner (2024)

Published in *Electronic Markets*, this paper conceptualizes Digital Product Passports (DPPs) for a low-carbon hydrogen market to enhance trust in global supply chains. The authors argue that a lack of credible sustainability data is a major barrier for emerging hydrogen markets and propose DPPs as a mechanism for verifying sustainability claims across complex, globalized supply chains. Notably, Körner is a co-author of works already cited in the ICIS paper's literature review, placing this within the same research ecosystem. The DPP concept is complementary to the GO scheme approach: where GOs certify energy attributes at the point of production, DPPs trace those attributes through the supply chain.

**Relevance:** Directly addresses MR6 (Unreliable Data Sources) and MR7 (Lack of Transparency). The DPP framing strengthens the argument for DP2 (Verifiability Through Transparency) and suggests that the ICIS paper's prototype could be extended toward DPP functionality for hydrogen supply chains.

### Mould, Silva & Knott (2022)

This *International Journal of Hydrogen Energy* paper provides a comparative analysis of biogas and hydrogen certification systems, examining how blockchain could guarantee environmentally positive origins. The authors note that production of these gases is not always as transparent as perceived, and propose certification combined with blockchain as a solution. This directly maps to the ICIS paper's cross-carrier scenario: converting GOs from electricity to hydrogen or biogas requires exactly the kind of cross-scheme data that Mould et al. identify as lacking transparency.

**Relevance:** Directly validates MR2 (Harmonization across carriers) and MR5 (Temporal/Spatial Decoupling). The cross-carrier comparison (biogas vs. hydrogen) mirrors the ICIS paper's conversion issuance problem.

### Ferraro, Gallo & Ippolito (2023)

Ferraro et al. present a practical test bench for green hydrogen production with blockchain-based traceability and certification at IEEE EEEIC. The prototype demonstrates that blockchain characteristics can solve traceability challenges for green hydrogen, contributing to its diffusion and market creation. As a proof-of-concept implementation, it provides empirical evidence that DLT-based traceability is technically feasible for hydrogen certification — corroborating the ICIS paper's Caliper benchmark findings from a different angle.

**Relevance:** Supports MR6 (Unreliable Data Sources) and DP2 (Verifiability Through Transparency). Provides independent validation that blockchain-based hydrogen certification prototypes are technically viable.

### Bhavana, Anand & Ramprabhakar (2024)

This *Scientific Reports* topical review covers blockchain applications in P2P energy markets and green hydrogen supply chains. The review identifies four primary application areas: P2P trading, green hydrogen supply chains, demand response, and REC tracking. The breadth of this review confirms that the ICIS paper operates at an intersection of multiple active research threads, strengthening its positioning as an integrative contribution.

**Relevance:** Supports MR3 (High Transaction Costs) and the general case for blockchain in energy certificate management. Confirms the research gap at the intersection of P2P markets and hydrogen certification.

### Stream A — Synthesis

The five papers in this stream collectively confirm that DLT-based certification for hydrogen and biogas is an active and growing research area. However, none address the *cross-carrier conversion issuance* problem — the mechanism by which a GO for one energy carrier (e.g., electricity) is converted into a GO for another (e.g., hydrogen) while preserving data integrity across separate scheme registries. This gap validates the ICIS paper's core contribution.

---

## 3. Stream B — Blockchain for Renewable Energy Certificates

This stream covers systems for issuing, trading, and managing RECs/GOs on blockchain platforms, with emphasis on privacy, tokenization, and market design.

### Liu, Chiu & Hua (2023)

Published in *Energy*, this paper proposes a blockchain-enabled REC trading system with a privacy-preserving approach. The privacy focus is highly relevant to the ICIS paper's DP3 (Privacy Through Selective Disclosure): Liu et al. address the tension between transparency of certificate provenance and confidentiality of commercial trading data — the same tension that motivates the use of Hyperledger Fabric's private data collections in the ICIS prototype.

**Relevance:** Directly supports MR8 (Invasion of Privacy) and validates the design rationale behind DP3. Offers an alternative privacy mechanism to compare against the ICIS paper's PDC-based approach.

### Zuo (2022)

Zuo proposes a blockchain-based decentralized platform for REC issuance and trading in *IEEE Access*. The core argument is that existing REC systems are centralized, highly regulated, and operationally expensive. Tokenizing RECs as blockchain tokens ensures trustworthy, immutable records while reducing operational costs. This mirrors the ICIS paper's argument about centralized GO registries (C1–C3) and supports the case for MR3 (High Transaction Costs) and DP1 (Scalability Through Automation). However, Zuo does not consider cross-carrier scenarios or conversion issuance.

**Relevance:** Validates MR1 (Low Incentivization) and MR3 (High Transaction Costs). Demonstrates independent convergence on tokenization as a design approach for energy certificates.

### Ferdous, Cali & Halden (2023)

This *Journal of Cleaner Production* paper leverages self-sovereign identity (SSI) and DLT for REC ecosystems. The SSI approach is a distinct architectural alternative to the ICIS paper's PDC-based privacy model, offering identity-layer privacy rather than data-layer privacy. Both address MR8 (Privacy) but through different mechanisms: SSI controls *who can identify participants*, while PDCs control *who can see transaction data*.

**Relevance:** Offers a complementary privacy architecture to the ICIS paper. Supports MR8, MR10 (Single Point of Failure), DP3 (Privacy), and DP6 (Governance Through Distributed Control).

### Cali, Kuzlu & Sebastian-Cardenas (2022)

Published in *Electrical Engineering*, this paper presents a cybersecure and scalable token-based REC framework. The focus on both security and scalability aligns with the ICIS paper's dual concern of DP4 (Security Through Decentralization) and DP1 (Scalability). As one of several token-based REC systems proposed in this period, it confirms an emerging consensus that tokenization is the appropriate abstraction for digital energy certificates.

**Relevance:** Supports MR3 (Transaction Costs) and DP1 (Scalability). Validates the token-based certificate representation adopted in the ICIS paper.

### Eickhoff, Exner & Busboom (2023)

Eickhoff et al. propose Energy Consumption Tokens for end-to-end trading of green energy certificates at IEEE GTD. Critically, they extend the Energy Web Origin platform by tying energy attribute certificates to both generation *and consumption* points — addressing a gap where most systems only track generation. This is relevant to the ICIS paper's device attestation mechanism (v6.0–v7.0), which similarly links certificates to metering devices at both ends of the energy conversion.

**Relevance:** Supports MR4 (Complexity/Manual Execution) and MR5 (Temporal/Spatial Decoupling). The consumption-side linking mirrors the ICIS paper's device-to-certificate attestation approach.

### Bjørn, Lloyd & Brander (2022)

Published in *Nature Climate Change* (134 cites), Bjørn et al. argue that renewable energy certificates threaten the integrity of corporate science-based targets. This high-impact paper provides authoritative evidence that the current centralized REC system has fundamental integrity problems — exactly the motivation behind MR9 (Double Counting) and MR7 (Lack of Transparency). The paper strengthens the ICIS paper's problem relevance by demonstrating that even the existing single-carrier REC system suffers from credibility issues; the cross-carrier extension makes these problems worse without the kind of DLT-based verification the ICIS paper proposes.

**Relevance:** Strongly supports MR7 (Transparency) and MR9 (Double Counting). Provides high-impact evidence for the urgency of the ICIS paper's problem statement.

### Stream B — Synthesis

The REC/GO blockchain literature shows strong convergence on three design choices: (1) tokenization of certificates, (2) smart-contract automation of issuance/retirement, and (3) some form of privacy mechanism. However, none of these systems address multi-carrier conversion or cross-registry interoperability. The ICIS paper sits at the frontier where these single-carrier blockchain-REC systems must be connected across carrier boundaries.

---

## 4. Stream C — Hyperledger Fabric Performance and Privacy

This stream covers HLF-specific research relevant to the ICIS paper's platform choice and performance claims.

### Guggenberger, Sedlmeir & Fridgen (2022)

This *Computers & Industrial Engineering* paper by the FIT research group (whose members are already heavily cited in the ICIS paper) presents an in-depth performance analysis of Hyperledger Fabric. The authors address the scarcity of empirical data on HLF's real-world performance — precisely the gap the ICIS paper fills for the GO domain. Key findings on MVCC contention, endorsement policy overhead, and CouchDB vs. LevelDB trade-offs are directly applicable to the ICIS prototype's design decisions (ADR-001 on hash-based IDs to avoid MVCC contention, CouchDB choice for rich queries).

**Relevance:** Foundational for DP1 (Scalability) — the performance characteristics documented here inform the ICIS paper's optimization strategy. The 83 citations confirm this is the reference work for HLF performance analysis.

### Tkachuk, Ilie & Robert (2023a)

Published in *Sustainable Energy, Grids and Networks*, this paper describes an HLF-based P2P renewable energy marketplace using private data collections for privacy. The authors explicitly use Fabric PDCs to separate public market operations from private trading data — the same architectural pattern the ICIS paper employs for separating public GO metadata from private conversion issuance details. This is the most architecturally similar system to the ICIS prototype found in the recent literature.

**Relevance:** Strongly validates DP3 (Privacy Through Selective Disclosure) and the choice of HLF PDCs. Independently confirms that the PDC pattern is appropriate for energy systems requiring both transparency and confidentiality.

### Tkachuk, Ilie & Robert (2023b)

The companion paper in *Annals of Telecommunications* benchmarks consensus mechanism performance in privacy-enabled energy marketplaces. The scalability findings complement the ICIS paper's Caliper benchmarks by examining consensus-level bottlenecks that affect all HLF-based energy systems. The documented throughput limitations align with the ICIS paper's v7.0 findings and support the architectural motivation for the v8.0 multi-channel design.

**Relevance:** Supports DP1 (Scalability) with independent performance data. Validates the ICIS paper's observation that single-channel HLF architectures face scalability ceilings.

### Pradhan, Singh & Verma (2022)

Published in *Scientific Reports*, this paper proposes a lightweight P2P energy trading framework using IOTA Tangle rather than traditional blockchain, achieving higher throughput for micro-transactions. While the platform choice differs (IOTA vs. HLF), the paper highlights the performance limitations of HLF for high-volume energy trading — supporting the ICIS paper's v8.0 redesign rationale.

**Relevance:** Provides a contrasting platform approach that validates MR3 (Transaction Costs). The throughput comparison strengthens the case for the ICIS paper's multi-channel v8.0 architecture as a necessary scalability measure.

### Stream C — Synthesis

The HLF literature confirms two key findings that underpin the ICIS paper's design: (1) PDCs are an effective privacy mechanism for energy applications (Tkachuk et al. 2023a), and (2) single-channel HLF architectures face real scalability limits (Guggenberger et al. 2022; Tkachuk et al. 2023b). The ICIS paper's progression from v7.0 (single-channel) to v8.0 (multi-channel) is well-grounded in this evidence base.

---

## 5. Stream D — Double Counting, Residual Mix, and Multi-Carrier Systems

### Holzapfel, Bach & Finkbeiner (2023)

This *International Journal of Life Cycle Assessment* paper systematically analyzes double counting challenges in electricity accounting for LCA. The authors identify inconsistent application of location-based vs. market-based methods as a root cause of double counting — precisely the mechanism that motivates MR9 in the ICIS paper. Their proposed remedies (clear method alignment, residual mix correction) are what the ICIS paper automates through smart-contract enforcement.

**Relevance:** Directly validates MR9 (Double Counting) with quantitative evidence. Provides the accounting-theory foundation for DP2 (Verifiability Through Transparency).

### Holzapfel, Bunsen & Schmidt-Sierra (2024)

The follow-up paper quantifies the effects of replacing location-based with market-based residual mixes across the entire ecoinvent database. The analysis reveals that double counting of renewable electricity leads to systematic impact underestimation. This strengthens the ICIS paper's argument that double counting in the GO system is not a theoretical concern but a measurable, consequential problem.

**Relevance:** Provides quantitative evidence for MR9 (Double Counting) severity. Strengthens the practical relevance argument.

### Marin, Cioara & Anghel (2023)

Published in *Future Internet*, this paper proposes blockchain-based multi-energy flexibility trading for buildings using ERC-1155 multi-token standards. Buildings are positioned as active flexibility assets within integrated multi-energy grids, trading both heat and electricity in community-level marketplaces. The multi-carrier framing (heat + electricity) and the focus on interoperability between energy types make this the closest analogy to the ICIS paper's cross-carrier GO conversion problem at the building/district scale.

**Relevance:** Supports MR2 (Harmonization) and DP5 (Interoperability Through Standardized Protocols). The multi-token approach to representing different energy carriers is an alternative design to the ICIS paper's single-asset-type-with-carrier-field model.

### Jansons, Zemīte & Zeltins (2022)

Published in *Latvian Journal of Physics and Technical Sciences*, this paper examines green hydrogen within the EU's gaseous fuel diversification strategy. The authors frame hydrogen as a key decarbonization vector requiring sector coupling — the integration of previously separate energy sectors (electricity, gas, heat). This sector-coupling context is precisely the regulatory environment in which the ICIS paper's cross-carrier GO system must operate: RED III mandates GOs not just for electricity but across all renewable energy carriers.

**Relevance:** Supports MR2 (Harmonization) and MR5 (Temporal/Spatial Decoupling). Provides the EU energy policy context motivating the ICIS paper's problem statement.

### Stream D — Synthesis

The double counting and multi-carrier literature provides strong quantitative and regulatory evidence for the urgency of the ICIS paper's problem. Holzapfel et al.'s work demonstrates that double counting is measurable and consequential; Marin et al. show that multi-carrier energy trading requires explicit interoperability mechanisms; and the hydrogen policy literature (Jansons et al.) confirms the EU-level regulatory mandate driving cross-carrier GO schemes.

---

## 6. Stream E — Supporting Context

### Sedlmeir, Lautenschlager & Fridgen (2022)

This *Electronic Markets* position paper (142 cites) discusses the transparency challenges of blockchain in organizations. The authors argue that blockchain's information exposure goes well beyond GDPR conflicts and illustrate the trade-off between protecting sensitive information and increasing process efficiency through smart contracts. They explore how permissioned blockchains and cryptographic technologies (including SSI and zero-knowledge proofs) can mitigate excessive transparency. This directly informs the ICIS paper's DP3: Fabric PDCs are one such mitigation mechanism, controlling data visibility within a permissioned network.

**Relevance:** Provides theoretical grounding for MR7 (Transparency) and MR8 (Privacy). Validates the tension between transparency and privacy that motivates DP2 and DP3 as complementary principles.

### Juszczyk & Shahzad (2022)

This *Energies* review covers blockchain principles, applications, and prospects for renewable energy. As a survey with 88 citations, it serves as a reference point for the state of blockchain-for-energy research at the start of the ICIS paper's development period (2022). The review identifies decentralization, IoT integration, and smart contracting as primary value propositions — all present in the ICIS paper's prototype design.

**Relevance:** Provides general domain context for MR1–MR4. Confirms that the ICIS paper operates within a recognized and active research domain.

### Nour, Chaves Ávila & Sánchez Miralles (2022)

This *IEEE Access* review (58 cites) examines blockchain potential in the electricity sector with emphasis on challenges for large-scale adoption. The identified barriers — scalability, regulatory uncertainty, integration with legacy systems — directly map to the challenges (C1–C8) the ICIS paper addresses. The review's focus on metering, billing, and P2P trading applications provides context for the ICIS paper's device attestation and oracle components (v6.0–v7.0).

**Relevance:** Supports MR3 (Transaction Costs) and DP1 (Scalability). Highlights the same large-scale adoption barriers that motivate the v8.0 multi-channel redesign.

### Onukwulu, Odochi-Agho & Eyo-Udo (2025)

This recent review in the *International Journal of Research and Innovation in Applied Science* covers blockchain integration for transparent renewable energy supply chains. The paper discusses energy certificates, EACs, and double counting within a blockchain-enabled supply chain context. As a 2025 publication, it confirms that the themes of the ICIS paper remain current and that the research gap (cross-carrier interoperability) persists.

**Relevance:** Confirms continued relevance of MR7 (Transparency) and MR9 (Double Counting) topics in 2025. Validates the ICIS paper's timeliness.

### Stream E — Synthesis

The supporting literature confirms that the ICIS paper operates within an active, well-recognized research domain. The transparency–privacy tension (Sedlmeir et al. 2022), adoption barriers (Nour et al. 2022), and continued research interest (Onukwulu et al. 2025) provide a solid contextual foundation.

---

## 7. Cross-Cutting Findings

### 7.1 Convergence on Design Choices

Across the 23 papers, several design choices emerge as near-consensus:

| Design Choice | Supporting Papers | ICIS Paper Alignment |
|---------------|-------------------|----------------------|
| **Tokenization** of certificates | Zuo 2022, Cali et al. 2022, Eickhoff et al. 2023, Marin et al. 2023 | ✓ GO assets as on-chain tokens |
| **Smart-contract automation** of issuance/retirement | Zuo 2022, Cali et al. 2022, Bhavana et al. 2024 | ✓ Chaincode-automated lifecycle |
| **Privacy mechanisms** for commercial data | Liu et al. 2023, Ferdous et al. 2023, Tkachuk et al. 2023a, Sedlmeir et al. 2022 | ✓ Fabric PDCs (DP3) |
| **Permissioned blockchain** for regulated markets | Guggenberger et al. 2022, Tkachuk et al. 2023a/b, Schmid et al. 2024 | ✓ HLF (permissioned by design) |

### 7.2 Identified Research Gap

No paper in the corpus addresses the **cross-carrier conversion issuance** problem — i.e., the mechanism by which a GO for one energy carrier is converted into a GO for another carrier while preserving verifiable data integrity across separate scheme registries. This gap is confirmed across all five streams:

- **Stream A** (Schmid et al. 2024, Mould et al. 2022): analyse hydrogen or biogas certification *individually*, not cross-carrier conversion.
- **Stream B** (Zuo 2022, Liu et al. 2023, etc.): focus on single-carrier REC systems.
- **Stream C** (Guggenberger et al. 2022, Tkachuk et al. 2023a/b): address HLF performance/privacy generically, not GO-specific cross-domain sharing.
- **Stream D** (Holzapfel et al. 2023/2024): identify double counting problems without proposing DLT-based solutions.
- **Stream E** (general reviews): note interoperability as an open challenge but do not propose concrete architectures.

### 7.3 Challenge–Paper Mapping

| Challenge | Supporting New Papers |
|-----------|----------------------|
| C1–C3 (Centralization, Cost, Manual processes) | Zuo 2022, Cali et al. 2022, Nour et al. 2022, Pradhan et al. 2022 |
| C4 (Lack of multi-carrier interoperability) | Marin et al. 2023, Mould et al. 2022 — *partially; gap persists* |
| C5 (Privacy vs. Transparency) | Liu et al. 2023, Sedlmeir et al. 2022, Tkachuk et al. 2023a, Ferdous et al. 2023 |
| C6 (Double counting) | Bjørn et al. 2022, Holzapfel et al. 2023, Holzapfel et al. 2024 |
| C7 (Scalability) | Guggenberger et al. 2022, Tkachuk et al. 2023b, Pradhan et al. 2022 |
| C8 (Single point of failure) | Ferdous et al. 2023, Schmid et al. 2024 |

---

## 8. Conclusion

The 23 papers analysed here confirm that the ICIS paper's problem domain — DLT-based energy certificate systems — is an active research area with strong convergence on core design choices (tokenization, smart-contract automation, privacy mechanisms, permissioned platforms). The literature provides independent support for each of the paper's meta-requirements and design principles. Crucially, the systematic search confirms a **persistent research gap** at the intersection of cross-carrier energy conversion and DLT-based certificate management: no existing work addresses the conversion issuance problem that sits at the heart of the ICIS contribution. This positions the paper as a genuine extension of the knowledge frontier rather than an incremental addition to an already-solved problem space.

Published in *Future Internet*, this paper proposes a blockchain solution for multi-energy flexibility trading in buildings using ERC-1155 multi-token standards. The multi-carrier approach (heat + electricity) within a single blockchain platform addresses the interoperability challenge from a different angle: while the ICIS paper connects separate registries across carrier boundaries, Marin et al. design a unified platform for multi-carrier trading. The ERC-1155 multi-token standard is an alternative token design to the ICIS paper's per-carrier namespaces.

**Relevance:** Supports MR2 (Harmonization) and DP5 (Interoperability). The multi-token approach offers a design alternative to compare against the ICIS paper's namespace-based carrier separation.

### Jansons, Zemīte & Zeltins (2022)

This *Latvian Journal of Physics and Technical Sciences* paper analyzes green hydrogen and EU gaseous fuel diversification risks, including sector coupling implications. The regulatory analysis of hydrogen's role in EU energy diversification provides context for the ICIS paper's RED III-driven problem statement. The sector coupling dimension confirms that cross-carrier energy systems are a policy priority, not merely a technical curiosity.

**Relevance:** Supports MR2 (Harmonization) and the overall problem relevance argument. Provides regulatory context for cross-carrier GO schemes.

### Stream D — Synthesis

The double counting literature (Holzapfel et al. 2023, 2024) provides quantitative evidence for one of the ICIS paper's core motivating challenges. Bjørn et al. (2022) in Nature Climate Change adds high-impact validation. Together, these papers strengthen the argument that DLT-based verification (as proposed in the ICIS paper) is necessary to prevent double counting in an increasingly complex, multi-carrier GO market.

---

## 6. Stream E — Broader Context

### Sedlmeir, Lautenschlager & Fridgen (2022)

Published in *Electronic Markets* (142 cites), this position paper discusses the transparency challenge of blockchain in organizations: the tension between excessive transparency (all participants see all data) and the need for privacy. The authors explore permissioned blockchains and cryptographic techniques (including SSI and zero-knowledge proofs) as solutions. This paper provides the theoretical underpinning for the ICIS paper's privacy design: the same trade-off between transparency and confidentiality drives DP3 (Privacy Through Selective Disclosure) and the choice of Fabric's private data collections.

**Relevance:** Foundational for MR7 (Transparency) and MR8 (Privacy). Provides the theoretical framing for the transparency-privacy balance that DP2 and DP3 jointly address.

### Juszczyk & Shahzad (2022)

This *Energies* review surveys blockchain technology for renewable energy, covering decentralization, IoT integration, and smart contracting. As a broad survey (88 cites), it confirms that the ICIS paper's research area is well-established and active. The review identifies maturity gaps in blockchain-for-energy applications — gaps the ICIS paper addresses with its production-grade prototype.

**Relevance:** Provides general support for MR1–MR4. Confirms the research area's relevance and the gap between conceptual proposals and production-ready implementations.

### Nour, Chaves Ávila & Sánchez Miralles (2022)

This *IEEE Access* review catalogs blockchain applications in the electricity sector and challenges for large-scale adoption. The challenges identified (scalability, interoperability, regulatory alignment) map directly to the ICIS paper's C1–C8. The emphasis on large-scale adoption barriers provides additional motivation for the ICIS paper's multi-version benchmarking approach.

**Relevance:** Supports MR3 (Transaction Costs) and DP1 (Scalability). The adoption challenges reinforce the practical relevance of the ICIS paper's performance evaluation.

### Onukwulu et al. (2025)

This recent review covers advances in blockchain integration for transparent renewable energy supply chains, including energy certificate tracking, traceability, and smart-contract automation. As a 2025 publication, it confirms that the ICIS paper's topic remains at the frontier of current research.

**Relevance:** Supports MR7 (Transparency) and MR9 (Double Counting). Confirms ongoing research interest.

---

## 7. Cross-Stream Synthesis and Implications

### 7.1 Convergent Findings

Across all 23 papers, several consistent themes emerge that support the ICIS paper's design decisions:

| Theme | Supporting Papers | ICIS Paper Alignment |
|-------|-------------------|---------------------|
| Tokenization of energy certificates | Zuo 2022; Cali et al. 2022; Eickhoff et al. 2023; Marin et al. 2023 | Consistent with GO-as-asset design in chaincode |
| PDC-based privacy for energy systems | Tkachuk et al. 2023a; Liu et al. 2023 | Validates DP3 (Privacy Through Selective Disclosure) |
| HLF scalability limits | Guggenberger et al. 2022; Tkachuk et al. 2023b; Pradhan et al. 2022 | Supports v8.0 multi-channel redesign motivation |
| Double counting as systemic risk | Bjørn et al. 2022; Holzapfel et al. 2023, 2024 | Validates MR9 and the need for automated verification |
| Hydrogen certification need | Schmid et al. 2024; Heeß et al. 2024; Ferraro et al. 2023 | Confirms problem relevance for cross-carrier GOs |

### 7.2 Identified Research Gap

None of the 23 papers address the **cross-carrier conversion issuance problem** — the mechanism by which a GO for one energy carrier is converted to a GO for another while maintaining data integrity across separate scheme registries. Individual papers address: single-carrier REC/GO blockchain systems (Stream B), hydrogen certification (Stream A), HLF privacy mechanisms (Stream C), and double counting risks (Stream D). But the integration point — conversion issuance as a cross-domain data sharing problem — remains unaddressed. This confirms the ICIS paper's contribution as a genuine gap-fill.

### 7.3 Design Principle Validation

| Design Principle | External Evidence |
|------------------|-------------------|
| DP1 (Scalability Through Automation) | Guggenberger et al. 2022 confirm MVCC contention as key bottleneck; Zuo 2022 and Cali et al. 2022 validate automation of certificate lifecycle |
| DP2 (Verifiability Through Transparency) | Bjørn et al. 2022 demonstrate integrity failures in current systems; Holzapfel et al. 2023/2024 quantify double counting risks |
| DP3 (Privacy Through Selective Disclosure) | Tkachuk et al. 2023a independently validate PDC pattern for energy; Liu et al. 2023 offer alternative privacy approach; Sedlmeir et al. 2022 provide theoretical framing |
| DP4 (Security Through Decentralization) | Cali et al. 2022 emphasize cybersecurity of token-based frameworks; Schmid et al. 2024 validate DLT for trustless information sharing |

### 7.4 Potential Extensions Suggested by the Literature

1. **Digital Product Passport integration** (Heeß et al. 2024): The prototype's GO data model could be extended toward DPP standards for hydrogen supply chain traceability.
2. **Self-sovereign identity layer** (Ferdous et al. 2023): Adding SSI on top of the PDC architecture would provide identity-layer privacy in addition to data-layer privacy.
3. **Consumption-side attestation** (Eickhoff et al. 2023): The v7.0 device attestation mechanism could be extended to certify energy consumption, enabling full end-to-end certificate tracking.
4. **Multi-token standards** (Marin et al. 2023): ERC-1155-style multi-tokens could simplify cross-carrier representation within a single smart contract.

---

## 8. Summary

The analysis of 23 recent papers (2022–2025) confirms the ICIS paper's problem relevance, validates its core design decisions, and identifies a clear research gap that the paper fills. The cross-carrier conversion issuance problem — requiring cross-domain data sharing between GO scheme registries — is not addressed by any of the recent publications reviewed. Existing work covers single-carrier blockchain-REC systems, hydrogen certification, HLF privacy mechanisms, and double counting, but does not integrate these concerns at the cross-carrier boundary. The ICIS paper's design principles find independent support in the recent literature, particularly DP3 (PDC-based privacy, validated by Tkachuk et al. 2023a) and DP1 (scalability, grounded in Guggenberger et al. 2022's HLF performance analysis).

---

## References (23 papers analyzed)

1. Bhavana, G.B. & Anand, R.S. (2024). Applications of blockchain technology in peer-to-peer energy markets and green hydrogen supply chains. *Scientific Reports*. DOI: 10.1038/s41598-024-72642-2
2. Bjørn, A., Lloyd, S.M. & Brander, M. (2022). Renewable energy certificates threaten the integrity of corporate science-based targets. *Nature Climate Change*. DOI: 10.1038/s41558-022-01379-5
3. Cali, Ü., Kuzlu, M. & Sebastian-Cardenas, D.J. (2022). Cybersecure and scalable, token-based renewable energy certificate framework. *Electrical Engineering*. DOI: 10.1007/s00202-022-01688-0
4. Eickhoff, M., Exner, A. & Busboom, A. (2023). Energy Consumption Tokens for Blockchain-Based End-to-End Trading of Green Energy Certificates. *IEEE GTD*. DOI: 10.1109/gtd49768.2023.00027
5. Ferdous, M.S., Cali, Ü. & Halden, U. (2023). Leveraging self-sovereign identity & DLT in renewable energy certificate ecosystems. *Journal of Cleaner Production*. DOI: 10.1016/j.jclepro.2023.138355
6. Ferraro, M. & Gallo, P. (2023). A test bench for the production of green hydrogen and its traceability and certification using blockchain technology. *IEEE EEEIC*. DOI: 10.1109/eeeic/icpseurope57605.2023.10194789
7. Guggenberger, T., Sedlmeir, J. & Fridgen, G. (2022). An in-depth investigation of the performance characteristics of Hyperledger Fabric. *Computers & Industrial Engineering*. DOI: 10.1016/j.cie.2022.108716
8. Heeß, P., Rockstuhl, J. & Körner, M.-F. (2024). Enhancing trust in global supply chains: Conceptualizing Digital Product Passports for a low-carbon hydrogen market. *Electronic Markets*. DOI: 10.1007/s12525-024-00690-7
9. Holzapfel, P., Bach, V. & Finkbeiner, M. (2023). Electricity accounting in life cycle assessment: the challenge of double counting. *Int J Life Cycle Assessment*. DOI: 10.1007/s11367-023-02158-w
10. Holzapfel, P., Bunsen, J. & Schmidt-Sierra, I. (2024). Replacing location-based electricity consumption with market-based residual mixes. *Int J Life Cycle Assessment*. DOI: 10.1007/s11367-024-02294-x
11. Jansons, L., Zemīte, L. & Zeltins, N. (2022). The Green Hydrogen and the EU Gaseous Fuel Diversification Risks. *Latvian J Physics & Technical Sciences*. DOI: 10.2478/lpts-2022-0033
12. Juszczyk, O. & Shahzad, K. (2022). Blockchain Technology for Renewable Energy: Principles, Applications and Prospects. *Energies*. DOI: 10.3390/en15134603
13. Liu, W.-J., Chiu, W.-Y. & Hua, W. (2023). Blockchain-enabled renewable energy certificate trading: A secure and privacy-preserving approach. *Energy*. DOI: 10.1016/j.energy.2023.130110
14. Marin, O., Cioara, T. & Anghel, I. (2023). Blockchain Solution for Buildings' Multi-Energy Flexibility Trading Using Multi-Token Standards. *Future Internet*. DOI: 10.3390/fi15050177
15. Mould, K. & Silva, F. (2022). A comparative analysis of biogas and hydrogen, and the impact of the certificates and blockchain new paradigms. *Int J Hydrogen Energy*. DOI: 10.1016/j.ijhydene.2022.09.107
16. Nour, M., Chaves Ávila, J.P. & Sánchez Miralles, Á. (2022). Review of Blockchain Potential Applications in the Electricity Sector. *IEEE Access*. DOI: 10.1109/access.2022.3171227
17. Onukwulu, E.C. et al. (2025). Advances in Blockchain Integration for Transparent Renewable Energy Supply Chains. *Int J Research and Innovation in Applied Science*. DOI: 10.51584/ijrias.2024.912058
18. Pradhan, N.R. & Singh, A.P. (2022). A blockchain based lightweight peer-to-peer energy trading framework. *Scientific Reports*. DOI: 10.1038/s41598-022-18603-z
19. Schmid, J., Ubacht, J. & van Engelenburg, S. (2024). Is it green? Designing a blockchain-based certification system for the EU hydrogen market. *Frontiers in Blockchain*. DOI: 10.3389/fbloc.2024.1408743
20. Sedlmeir, J., Lautenschlager, J. & Fridgen, G. (2022). The transparency challenge of blockchain in organizations. *Electronic Markets*. DOI: 10.1007/s12525-022-00536-0
21. Tkachuk, R.-V., Ilie, D. & Robert, R. (2023a). Towards efficient privacy and trust in decentralized blockchain-based peer-to-peer renewable energy marketplace. *Sustainable Energy, Grids and Networks*. DOI: 10.1016/j.segan.2023.101146
22. Tkachuk, R.-V., Ilie, D. & Robert, R. (2023b). On the performance and scalability of consensus mechanisms in privacy-enabled decentralized renewable energy marketplace. *Annals of Telecommunications*. DOI: 10.1007/s12243-023-00973-8
23. Zuo, Y. (2022). Tokenizing Renewable Energy Certificates (RECs)—A Blockchain Approach for REC Issuance and Trading. *IEEE Access*. DOI: 10.1109/access.2022.3230937
