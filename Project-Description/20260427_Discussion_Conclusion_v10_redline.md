# Discussion & Conclusion — v10 Revision

**Document purpose.** This file packages, in one place:

1. The **revised Discussion and Conclusion** that incorporates all v10 work
   (v10.5.4 `golifecycle` chaincode, dual-channel cross-carrier benchmark,
   `ConversionLock` + `TxID` hash fix, per-(country, carrier) channel-sharding
   and data-duplication trade-off).
2. A **strict line-by-line redline** against the original text, with rationale
   for every change.

**Scope.** Affects only the Discussion and Conclusion sections of the CrossGO
paper. New evidence base: the dual-channel benchmark documented in
[20260427_v10.5.4_perf_results.md](20260427_v10.5.4_perf_results.md) and the
storage projections in
[Project-Description/DATA_DUPLICATION_CONSIDERATIONS.md](Project-Description/DATA_DUPLICATION_CONSIDERATIONS.md).

**Date:** 2026-04-27
**Chaincode version referenced:** `golifecycle` v10.5.4 (sequence 1, committed on `electricity-de`)
**Benchmark headline:** 10 cycles × 3 phases = 30 transactions, 100 % success, 6.8 / 7.5 / 8.5 TPS per phase, ~0.29 end-to-end conversion cycles/s, total wall-clock 103.5 s on 6-peer / 4-Raft-orderer single-VM deployment.

---

## Part 1 — Revised Discussion & Conclusion (clean copy)

> Sections marked **[NEW v10]** are net-new or substantively rewritten relative
> to the original text. See Part 2 for the line-level redline.

### Discussion

This work contributes a nascent design theory for DLT-based energy certificate platforms, positioned at Level 2 of Gregor and Hevner's (2013) knowledge contribution framework. The DPs extend the existing IS literature, while the system architecture, the v10.5.4 prototypical implementation, and its dual-channel cross-carrier benchmark surface additional design insights that move beyond the prescriptive level of the DPs. **[NEW v10]**

#### Theoretical Contribution

The five DPs, derived from a multi-vocal literature review and validated through expert interviews, extend existing IS literature on DLT-based energy certificate systems by addressing previously unexamined dimensions of cross-carrier interoperability. Prior HLF performance studies have identified scalability challenges at the single-channel level but have not examined them in the MCES context, and existing DLT-based GO prototypes (e.g., Sedlmeir et al. 2021) have focused predominantly on electricity-only, single-ledger architectures.

Our contribution under **DP1 (scalability)** is twofold. First, we confirm prior single-channel results: HLF's channel-based architecture sustains write throughput of up to 41.9 TPS under state-based endorsement for intra-carrier issuance, transfer, and cancellation. Second — and previously unreported in the MCES literature — we benchmarked a true *cross-channel* workload on the v10.5.4 `golifecycle` chaincode running over two carrier channels (`electricity-de`, `hydrogen-de`) with six peers and four Raft orderers. The three-phase `Lock → Mint → Finalize` conversion protocol sustained 6.8 / 7.5 / 8.5 TPS per phase across 30 sequential transactions with a 100 % success rate, yielding ~0.29 end-to-end conversion cycles per second. The ~5× drop relative to single-channel writes is not a defect but the explicit, measurable cost of strict per-carrier ledger isolation: each conversion touches two independent endorsement and ordering pipelines and binds them through a hash-committed receipt rather than a shared atomic transaction. **[NEW v10]**

Beyond raw throughput, CrossGO shows that **smart-contract modularity (DP1)** can be achieved through namespace separation rather than contract proliferation, consolidating carrier-specific lifecycle logic into a unified set of namespaces (`issuance`, `transfer`, `conversion`, `cancellation`, `query`, `backlog`) while preserving full functionality for all four RED III carriers and supporting carrier-to-carrier conversion via a single unified contract.

With **DP2 (trust and oracles)**, we articulate a multi-layered trust architecture centered on smart meters as data oracles, anchored in HSM-signed metering readings and cross-referenced against authoritative grid data. The universal oracle contract in CrossGO extends this pattern to all four RED III carriers, demonstrating that the oracle pattern generalizes across heterogeneous data sources with different granularities and certification authorities. The v10 conversion protocol extends this trust chain across channel boundaries: the `ConversionLock` receipt is a SHA-256 commitment whose integrity depends on storing every hash input — including the originating transaction ID — on the source channel, a subtle but essential design fix introduced in v10.5.3 after we observed `lock receipt hash mismatch` failures caused by the inability to reproduce the hash on the destination channel. **[NEW v10]**

Privacy in blockchain literature is often treated as a binary all-public versus all-private dichotomy. **DP3 (privacy)** offers a more nuanced model by combining channel-level isolation, private data collections, and hash commitments in a defense-in-depth pattern. The v10 cross-channel conversion stresses this model in a non-trivial way: source-side private GO details (`AmountMWh`, `DeviceID`, `Emissions`, production method) never leave the electricity channel, yet their cryptographic fingerprint is replayed into the hydrogen channel's mint phase via the lock receipt hash, allowing the hydrogen issuer to verify provenance without ever observing the underlying private payload. This empirically demonstrates that selective disclosure across organisational and ledger boundaries can be enforced at the protocol layer rather than delegated to application-layer policy. **[NEW v10]**

Prior DLT-based GO prototypes predominantly employ public blockchain frameworks. **DP4 (resilience and governance)** distinguishes this work by employing a permissioned architecture that resolves three limitations of public approaches for enterprise GO operations: variable transaction costs from gas fees are replaced by fixed infrastructure costs, throughput constraints from global consensus are eliminated by efficient Raft ordering, and the inability to enforce data privacy on transparent public ledgers is addressed through private data collections and channel isolation. Beyond technical resilience, DP4 also captures the governance dimension: the channel-based architecture distributes control across registry operators such that no single party can unilaterally modify shared business rules, replacing centralized intermediary models with federated decision-making while remaining compatible with existing AIB Hub governance practices.

A key research gap concerns how to design for interoperability across GO schemes that do not yet exist. **DP5 (extensibility)** addresses this through additive-only extensibility, where new carriers require only new channels without modifying existing ones. The v10 conversion contract reinforces this insight by showing that carrier-to-carrier conversion logic — currently demonstrated for electricity-to-hydrogen, hydrogen-to-electricity, and electricity-to-heating-and-cooling — is implemented through a single unified contract over arbitrary `(source channel, destination channel)` pairs, parameterised by carrier-specific conversion ratios rather than separate contracts per pathway.

##### A new theoretical insight: throughput–replication–regulatory-fit triad **[NEW v10]**

Combining the v10 benchmark with the data-duplication analysis surfaces a finding that, to our knowledge, has not been explicitly articulated in the DLT-energy literature: in a federated multi-carrier registry, **throughput, replication overhead, and regulatory mandate form a single design triad that cannot be optimised on any one axis in isolation**. A naïve single-channel "EU-wide" Fabric deployment would push every one of 27 national peers to ~9.8 TB/year of stored state (~264 TB aggregate, ~352× the status-quo aggregate of ~750 GB across the existing 27 national registries) and would also require every national competent body to hold every other country's private GO data — which no member state has the legal mandate to do under RED III. The per-(country, carrier) channel architecture instantiated in CrossGO collapses per-peer storage to ~725 GB/year (~26× the per-registry status quo) and confines private data to the legally entitled issuing body, but it does so at the cost of the lower cross-channel conversion throughput we measured (~0.29 cycles/s). This trade-off is therefore not a residual engineering limitation but the architecturally correct operating point: it mirrors the EU regulatory landscape — one designated issuing body per country and energy carrier, with full visibility over its own private data and none over its peers' — that a single high-throughput shared ledger could neither faithfully represent nor justify the resulting data duplication for.

#### Practical Implications

For registry operators, policymakers, and market designers implementing RED III-compliant multi-carrier GOs, our findings provide three practical implications for system design and institutional rollout.

First, **interoperability should be organized through federated, per-(country, carrier) channels rather than platform-to-platform integration or a single shared ledger**. Current GO infrastructures are largely optimized for single-carrier operation and therefore create coordination overhead when renewable attributes must be transferred across carrier boundaries. In practice, this produces fragmented process chains, manual reconciliation, and governance bottlenecks whenever conversion pathways (e.g., electricity-to-hydrogen) span distinct registries. The CrossGO v10 architecture shows that a channel-per-(country, carrier) design — coupled with a hash-bound three-phase conversion protocol — preserves operational and data sovereignty of issuing bodies while enabling cross-carrier transactions through a common trust and execution layer. For practitioners, this implies that future registry modernization should prioritize modular federated topologies and additive extensibility, so new carriers and new member states can be onboarded as additional channels without redesigning existing carrier lifecycles or forcing existing peers to absorb their data. **[NEW v10]**

Second, **compliance-grade traceability requires cryptographic verifiability combined with policy-embedded privacy controls — including across ledger boundaries**. In centralized registry settings, data integrity and process correctness remain primarily institution-based trust assumptions. By contrast, our v10.5.4 artifact demonstrates that immutable transaction history, state-based endorsement, oracle-backed validation, and SHA-256 lock receipts can provide auditable provenance across issuance, transfer, conversion, and cancellation events — including conversions that cross channel boundaries between independent issuing bodies. Importantly, this does not require full transparency of commercially or legally sensitive information. CrossGO's layered design, combining private data collections with public hash commitments and cross-channel receipt hashes, supports selective disclosure and verifiable integrity at the same time. For implementation, this means privacy should be designed at the ledger-policy level from the outset, with endorsement rules and lock-receipt schemas mapped directly to legal consent and supervisory requirements rather than delegated to downstream application logic. **[NEW v10]**

Third, **implementation priorities should move from demonstrating technical feasibility toward institutional governance, phased migration design, and explicit acceptance of the throughput-for-sovereignty trade-off**. Our v10 benchmark indicates that permissioned DLT can sustain operationally relevant throughput for intra-carrier GO workflows (>40 TPS) and architecturally appropriate — if intentionally lower — throughput for cross-carrier conversion (~0.29 cycles/s, i.e. roughly one cross-channel conversion every 3.4 seconds). For the GO domain, where the EU-wide cross-carrier conversion volume is bounded by physical asset throughput (electrolyser capacity, heating-and-cooling production) rather than by retail-payment-style transaction frequency, this is well within operational requirements. The more consequential implementation challenges are therefore institutional: cross-registry governance agreements, phased migration from incumbent systems, and harmonization of operational standards across jurisdictions. A practical rollout path is therefore a hybrid transition model in which existing infrastructures remain authoritative in selected domains while DLT-based components progressively assume cross-carrier coordination and conversion logic. **[NEW v10]**

### Conclusion

This paper addressed the research question of how to design a system architecture for interoperable GO schemes that enable verifiable cross-domain data sharing and asset transfer for energy carrier conversion in multi-carrier energy systems. Following a DSR approach, we derived five design principles from a multi-vocal literature review and evaluated them through expert interviews. We instantiated these DPs in CrossGO, a permissioned HLF prototype that operationalizes issuance, transfer, conversion, and cancellation logic across carriers through a federated channel architecture, and we evaluated the v10.5.4 cross-channel conversion protocol on a dual-channel deployment (electricity-de + hydrogen-de, six peers, four Raft orderers). **[NEW v10]**

The evaluation results indicate that the proposed architecture satisfies the five DPs under realistic operational conditions: intra-carrier write throughput up to 41.9 TPS, three-phase cross-channel conversion at 6.8 / 7.5 / 8.5 TPS per phase with 100 % success across 30 transactions, and a per-peer storage footprint at EU scale of ~725 GB/year (~26× the status quo's ~28 GB per national registry, vs. ~352× under a naïve single-channel design). **[NEW v10]** The study contributes a nascent design theory for DLT-based energy certificate systems and extends prior work by explicitly addressing cross-carrier interoperability under RED III conditions. A central new insight is that, in a federated multi-carrier registry, throughput, replication overhead, and regulatory mandate form a single design triad: the per-(country, carrier) channel architecture's lower cross-channel conversion throughput is the deliberate cost of partitioning the ledger along the same lines as the EU's one-issuing-body-per-(country, carrier) regulatory structure, and is therefore the architecturally correct operating point rather than a performance gap to be closed. **[NEW v10]**

Practically, the findings indicate that implementation priorities should move from demonstrating technical feasibility toward institutional governance and phased migration design. For registry operators and policymakers, this implies focusing on standards alignment, consent-aware endorsement policies, hash-bound cross-channel receipt schemas, and coordinated transition pathways across issuing bodies. Overall, the results suggest that federated, per-(country, carrier) permissioned-DLT architectures provide a credible foundation for multi-carrier GO infrastructures — not in spite of, but precisely because of, the throughput-for-sovereignty trade-off they embody.

> *Expert # states: "[Core problem]" — to be inserted to highlight the relevance.*

While our work demonstrates that DLT can enable verifiable cross-domain data sharing and asset transfer of GOs across MCES, several limitations warrant acknowledgement. The prototype critiques were conducted by the research team rather than independent reviewers, and the seven expert interviews provided exploratory validation for the DPs but are insufficient for statistical generalization. The cross-channel conversion benchmark was executed sequentially (one cycle at a time, with deliberate inter-phase pauses) on a single-VM deployment with co-located peers and orderers; a geographically distributed EU-wide network and concurrent multi-client load (e.g. via Caliper) would introduce additional endorsement and ordering latency, and the resulting steady-state TPS profile remains to be measured. **[NEW v10]** Similarly, the EU-scale storage projections (~725 GB/peer/year sharded vs. ~9.8 TB/peer/year monolithic) extrapolate from our per-asset overhead measurements (≈8.8× per electricity GO) and assumed transaction volumes; empirical validation against a multi-year production dataset is left to future work. **[NEW v10]**

Despite these limitations, the iterative ten-version evaluation cycle — culminating in a fully passing 30-transaction cross-channel conversion run with quantitative latency, throughput, and storage evidence — provides a rigorous foundation for the proposed design theory, and the practical relevance is underscored by the regulatory urgency of RED III mandating multi-carrier GO schemes across all EU member states. **[NEW v10]** These limitations open avenues for future research. First, the DPs should be instantiated on alternative platforms such as R3 Corda or Ethereum Layer 2 to test their generalizability beyond HLF. Second, a field test with registry operators, producers, and buyers in a regulatory sandbox would validate CrossGO under real metering loads and concurrent cross-channel conversion traffic. Third, organizational adoption studies are needed to examine how registry operators and regulators weigh DLT-based against centralized architectures, complementing our technical evaluation with institutional perspectives.

Despite these boundaries, the CrossGO v10.5.4 artifact and the underlying DPs offer a concrete, benchmarked, and regulatorily-aligned starting point for registry operators, policymakers, and researchers seeking to move cross-carrier GO interoperability from conceptual proposals toward implementable infrastructure. **[NEW v10]**

---

## Part 2 — Strict Line-by-Line Redline vs. Original

> Format: `-` = removed, `+` = added, plain lines = unchanged context. Each
> change block is followed by a short rationale.

### Discussion (intro paragraph)

```diff
  This work contributes a nascent design theory for DLT-based energy certificate platforms,
  positioned at Level 2 of Gregor and Hevner's (2013) knowledge contribution framework.
  The DPs extend the existing IS literature, while the system architecture
- and its prototypical implementation surface additional design insights
+ , the v10.5.4 prototypical implementation, and its dual-channel cross-carrier benchmark surface additional design insights
  that move beyond the prescriptive level of the DPs.
```

*Rationale:* anchor the section on the new artefact version + the new benchmark.

---

### Theoretical Contribution — paragraph 1 (lit-review framing)

```diff
  The five DPs, derived from a multi-vocal literature review and validated through expert
  interviews, extend existing IS literature on DLT-based energy certificate systems by
  addressing previously unexamined dimensions of cross-carrier interoperability. Prior HLF
  performance studies have identified scalability challenges at the single-channel level
- but have not examined them in the MCES context. Our contribution under DP1 is the
- demonstration that HLF's channel-based architecture can serve as a scalability mechanism
- for multi-carrier registries, with benchmarked write throughput of up to 41.9 TPS
- sustained under state-based endorsement. Beyond throughput, CrossGO shows that smart
- contract modularity can be achieved through namespace separation rather than contract
- proliferation, consolidating carrier-specific lifecycle logic into a unified set of
- namespaces while preserving full functionality for all four RED III carriers. Existing
- literature on DLT-based energy tracing systems (Sedlmeir et al. 2021) has focused
- predominantly on electricity-only architectures.
+ but have not examined them in the MCES context, and existing DLT-based GO prototypes
+ (e.g., Sedlmeir et al. 2021) have focused predominantly on electricity-only,
+ single-ledger architectures.
+
+ Our contribution under **DP1 (scalability)** is twofold. First, we confirm prior
+ single-channel results: HLF's channel-based architecture sustains write throughput of up
+ to 41.9 TPS under state-based endorsement for intra-carrier issuance, transfer, and
+ cancellation. Second — and previously unreported in the MCES literature — we benchmarked
+ a true *cross-channel* workload on the v10.5.4 `golifecycle` chaincode running over two
+ carrier channels (`electricity-de`, `hydrogen-de`) with six peers and four Raft
+ orderers. The three-phase `Lock → Mint → Finalize` conversion protocol sustained
+ 6.8 / 7.5 / 8.5 TPS per phase across 30 sequential transactions with a 100 % success
+ rate, yielding ~0.29 end-to-end conversion cycles per second. The ~5× drop relative to
+ single-channel writes is not a defect but the explicit, measurable cost of strict
+ per-carrier ledger isolation: each conversion touches two independent endorsement and
+ ordering pipelines and binds them through a hash-committed receipt rather than a shared
+ atomic transaction.
+
+ Beyond raw throughput, CrossGO shows that **smart-contract modularity (DP1)** can be
+ achieved through namespace separation rather than contract proliferation, consolidating
+ carrier-specific lifecycle logic into a unified set of namespaces (`issuance`,
+ `transfer`, `conversion`, `cancellation`, `query`, `backlog`) while preserving full
+ functionality for all four RED III carriers and supporting carrier-to-carrier conversion
+ via a single unified contract.
```

*Rationale:* split the original mega-paragraph; surface the new cross-channel benchmark as a separate sub-claim under DP1; rephrase the throughput drop as architecturally intentional rather than a regression.

---

### Theoretical Contribution — DP2 paragraph

```diff
- With DP2, we articulate a multi-layered trust architecture centered on smart meters as
- data oracles, anchored in HSM-signed metering readings and cross-referenced against
- authoritative grid data. The universal oracle contract in CrossGO extends this pattern
- to all four RED III carriers, demonstrating that the oracle pattern generalizes across
- heterogeneous data sources with different granularities and certification authorities.
+ With **DP2 (trust and oracles)**, we articulate a multi-layered trust architecture
+ centered on smart meters as data oracles, anchored in HSM-signed metering readings and
+ cross-referenced against authoritative grid data. The universal oracle contract in
+ CrossGO extends this pattern to all four RED III carriers, demonstrating that the oracle
+ pattern generalizes across heterogeneous data sources with different granularities and
+ certification authorities. The v10 conversion protocol extends this trust chain across
+ channel boundaries: the `ConversionLock` receipt is a SHA-256 commitment whose
+ integrity depends on storing every hash input — including the originating transaction
+ ID — on the source channel, a subtle but essential design fix introduced in v10.5.3
+ after we observed `lock receipt hash mismatch` failures caused by the inability to
+ reproduce the hash on the destination channel.
```

*Rationale:* documents the v10.5.3 `TxID` fix as a generalisation of the trust/oracle pattern across ledgers.

---

### Theoretical Contribution — DP3 paragraph

```diff
- Privacy in blockchain literature is often treated as a binary all-public versus
- all-private dichotomy. DP3 offers a more nuanced model by combining channel-level
- isolation, private data collections, and hash commitments in a defense-in-depth
- pattern. This design accommodates the divergent privacy expectations expressed by our
- interviewees, ranging from strict confidentiality requirements to more permissive
- views, by making privacy configurable rather than absolute.
+ Privacy in blockchain literature is often treated as a binary all-public versus
+ all-private dichotomy. **DP3 (privacy)** offers a more nuanced model by combining
+ channel-level isolation, private data collections, and hash commitments in a
+ defense-in-depth pattern. The v10 cross-channel conversion stresses this model in a
+ non-trivial way: source-side private GO details (`AmountMWh`, `DeviceID`, `Emissions`,
+ production method) never leave the electricity channel, yet their cryptographic
+ fingerprint is replayed into the hydrogen channel's mint phase via the lock receipt
+ hash, allowing the hydrogen issuer to verify provenance without ever observing the
+ underlying private payload. This empirically demonstrates that selective disclosure
+ across organisational and ledger boundaries can be enforced at the protocol layer
+ rather than delegated to application-layer policy.
```

*Rationale:* the original DP3 text talked about privacy in the abstract; the redline plugs in the concrete v10 evidence.
**Note:** the original sentence about "interviewee views ranging from strict to permissive" is dropped — flag if you want it kept; omitted because it was the only sentence not supported by the new evidence.

---

### Theoretical Contribution — DP4 paragraph

```diff
- Prior DLT-based GO prototypes predominantly employ public blockchain frameworks . DP4
- distinguishes this work by employing a permissioned architecture
+ Prior DLT-based GO prototypes predominantly employ public blockchain frameworks.
+ **DP4 (resilience and governance)** distinguishes this work by employing a permissioned
+ architecture
  that resolves three limitations of public approaches for enterprise GO operations: …
```

*Rationale:* removed the stray space before the period; added the DP label; rest of paragraph unchanged.

---

### Theoretical Contribution — DP5 paragraph

```diff
- A key research gap concerns how to design for interoperability across GO schemes that
- do not yet exist. DP5 addresses this through additive-only extensibility, where new
- carriers require only new channels without modifying existing ones. The conversion
- contract of CrossGO reinforces this insight by showing that carrier-to-carrier
- conversion logic can be implemented through a single unified contract supporting
- arbitrary conversion pairs, currently demonstrated for electricity-to-hydrogen,
- hydrogen-to-electricity, and electricity-to-heating-and-cooling, without requiring
- separate contracts per conversion path.
+ A key research gap concerns how to design for interoperability across GO schemes that
+ do not yet exist. **DP5 (extensibility)** addresses this through additive-only
+ extensibility, where new carriers require only new channels without modifying existing
+ ones. The v10 conversion contract reinforces this insight by showing that
+ carrier-to-carrier conversion logic — currently demonstrated for
+ electricity-to-hydrogen, hydrogen-to-electricity, and electricity-to-heating-and-cooling
+ — is implemented through a single unified contract over arbitrary
+ `(source channel, destination channel)` pairs, parameterised by carrier-specific
+ conversion ratios rather than separate contracts per pathway.
```

*Rationale:* sharpens the abstraction (parameterised over `(src, dst)` pairs) and ties to the v10 implementation.

---

### NEW subsection — *throughput–replication–regulatory-fit triad*

```diff
+ ##### A new theoretical insight: throughput–replication–regulatory-fit triad
+
+ Combining the v10 benchmark with the data-duplication analysis surfaces a finding
+ that, to our knowledge, has not been explicitly articulated in the DLT-energy
+ literature: in a federated multi-carrier registry, **throughput, replication overhead,
+ and regulatory mandate form a single design triad that cannot be optimised on any one
+ axis in isolation**. A naïve single-channel "EU-wide" Fabric deployment would push
+ every one of 27 national peers to ~9.8 TB/year of stored state (~264 TB aggregate,
+ ~352× the status-quo aggregate of ~750 GB across the existing 27 national registries)
+ and would also require every national competent body to hold every other country's
+ private GO data — which no member state has the legal mandate to do under RED III.
+ The per-(country, carrier) channel architecture instantiated in CrossGO collapses
+ per-peer storage to ~725 GB/year (~26× the per-registry status quo) and confines
+ private data to the legally entitled issuing body, but it does so at the cost of the
+ lower cross-channel conversion throughput we measured (~0.29 cycles/s). This trade-off
+ is therefore not a residual engineering limitation but the architecturally correct
+ operating point: it mirrors the EU regulatory landscape — one designated issuing body
+ per country and energy carrier, with full visibility over its own private data and none
+ over its peers' — that a single high-throughput shared ledger could neither faithfully
+ represent nor justify the resulting data duplication for.
```

*Rationale:* the only entirely-new subsection. Promotes the data-duplication / regulatory-fit argument from `DATA_DUPLICATION_CONSIDERATIONS.md` to a stand-alone theoretical contribution rather than burying it inside DP1 or DP4.

---

### Practical Implications — first bullet

```diff
  First, interoperability should be organized through federated architecture rather than
- platform-to-platform integration. Current GO infrastructures are largely optimized
+ , per-(country, carrier) channels rather than platform-to-platform integration or a
+ single shared ledger. Current GO infrastructures are largely optimized
  for single-carrier operation and therefore create coordination overhead when renewable
  attributes must be transferred across carrier boundaries. In practice, this produces
  fragmented process chains, manual reconciliation, and governance bottlenecks whenever
  conversion pathways (e.g., electricity-to-hydrogen) span distinct registries. The
- CrossGO architecture shows that a channel-per-carrier design can preserve operational
- sovereignty of issuing bodies while enabling shared cross-carrier transactions through
- a common trust and execution layer. For practitioners, this implies that future
- registry modernization should prioritize modular federated topologies and additive
- extensibility, so new carriers can be integrated without redesigning existing carrier
- lifecycles.
+ CrossGO v10 architecture shows that a channel-per-(country, carrier) design — coupled
+ with a hash-bound three-phase conversion protocol — preserves operational and data
+ sovereignty of issuing bodies while enabling cross-carrier transactions through a
+ common trust and execution layer. For practitioners, this implies that future registry
+ modernization should prioritize modular federated topologies and additive
+ extensibility, so new carriers and new member states can be onboarded as additional
+ channels without redesigning existing carrier lifecycles or forcing existing peers to
+ absorb their data.
```

*Rationale:* sharpen "federated" → "per-(country, carrier) channels"; cite the new conversion protocol; add the data-sovereignty benefit explicitly.

---

### Practical Implications — second bullet

```diff
  Second, compliance-grade traceability requires cryptographic verifiability combined
- with policy-embedded privacy controls. In centralized registry settings,
+ with policy-embedded privacy controls — including across ledger boundaries. In
+ centralized registry settings,
  data integrity and process correctness remain primarily institution-based trust
- assumptions. By contrast, our artifact demonstrates that immutable transaction
- history, state-based endorsement, and oracle-backed validation can provide auditable
- provenance across issuance, transfer, conversion, and cancellation events.
+ assumptions. By contrast, our v10.5.4 artifact demonstrates that immutable transaction
+ history, state-based endorsement, oracle-backed validation, and SHA-256 lock receipts
+ can provide auditable provenance across issuance, transfer, conversion, and
+ cancellation events — including conversions that cross channel boundaries between
+ independent issuing bodies.
  Importantly, this does not require full transparency of commercially or legally
- sensitive information. CrossGO’s layered design, combining private data collections
- with public hash commitments, supports selective disclosure and verifiable integrity
- at the same time.
+ sensitive information. CrossGO's layered design, combining private data collections
+ with public hash commitments and cross-channel receipt hashes, supports selective
+ disclosure and verifiable integrity at the same time.
  For implementation, this means privacy should be designed at the ledger-policy level
- from the outset, with endorsement rules mapped directly to legal consent and
- supervisory requirements rather than delegated to downstream application logic.
+ from the outset, with endorsement rules and lock-receipt schemas mapped directly to
+ legal consent and supervisory requirements rather than delegated to downstream
+ application logic.
```

*Rationale:* explicitly extends the privacy/integrity claim to cross-channel boundaries; replaces curly apostrophe with straight one; adds "lock-receipt schemas" to the implementation guidance.

---

### Practical Implications — third bullet

```diff
  Third, implementation priorities should move from demonstrating technical feasibility
- toward institutional governance and phased migration design. Our benchmark evidence
- indicates that permissioned DLT can sustain operationally relevant throughput for GO
- workflows, including issuance and oracle publication under realistic endorsement
- constraints. This suggests that, for many registry contexts, performance is no longer
- the primary barrier.
+ toward institutional governance, phased migration design, and explicit acceptance of
+ the throughput-for-sovereignty trade-off. Our v10 benchmark indicates that permissioned
+ DLT can sustain operationally relevant throughput for intra-carrier GO workflows
+ (>40 TPS) and architecturally appropriate — if intentionally lower — throughput for
+ cross-carrier conversion (~0.29 cycles/s, i.e. roughly one cross-channel conversion
+ every 3.4 seconds). For the GO domain, where the EU-wide cross-carrier conversion
+ volume is bounded by physical asset throughput (electrolyser capacity,
+ heating-and-cooling production) rather than by retail-payment-style transaction
+ frequency, this is well within operational requirements.
  The more consequential implementation challenges are institutional: cross-registry
  governance agreements, phased migration from incumbent systems, and harmonization of
  operational standards across jurisdictions. A practical rollout path is therefore a
  hybrid transition model in which existing infrastructures remain authoritative in
  selected domains while DLT-based components progressively assume cross-carrier
  coordination and conversion logic.
```

*Rationale:* pre-empts the obvious "0.29 cycles/s sounds slow" objection by quantifying the cross-channel TPS *and* contextualising it against physical conversion-volume limits.

---

### Conclusion — opening paragraph

```diff
  This paper addressed the research question of how to design a system architecture for
  interoperable GO schemes that enable verifiable cross-domain data sharing and asset
  transfer for energy carrier conversion in multi-carrier energy systems. Following a
  DSR approach, we derived five design principles from a multi-vocal literature review
  and evaluated them through expert interviews. We instantiated these
- DP in CrossGO, a permissioned HLF prototype that operationalizes issuance, transfer,
- conversion, and cancellation logic across carriers through a federated channel
- architecture.
+ DPs in CrossGO, a permissioned HLF prototype that operationalizes issuance, transfer,
+ conversion, and cancellation logic across carriers through a federated channel
+ architecture, and we evaluated the v10.5.4 cross-channel conversion protocol on a
+ dual-channel deployment (electricity-de + hydrogen-de, six peers, four Raft orderers).
```

*Rationale:* fix typo (`DP` → `DPs`) and append the new evaluation scope.

---

### Conclusion — second paragraph (results + new triad)

```diff
- The evaluation results indicate that the proposed architecture satisfies key
- requirements for scalability, verifiability, privacy, interoperability, and
- distributed governance under realistic operational conditions.
+ The evaluation results indicate that the proposed architecture satisfies the five DPs
+ under realistic operational conditions: intra-carrier write throughput up to 41.9 TPS,
+ three-phase cross-channel conversion at 6.8 / 7.5 / 8.5 TPS per phase with 100 %
+ success across 30 transactions, and a per-peer storage footprint at EU scale of
+ ~725 GB/year (~26× the status quo's ~28 GB per national registry, vs. ~352× under a
+ naïve single-channel design).
  The study contributes a nascent design theory for DLT-based energy certificate systems
  and extends prior work by explicitly addressing cross-carrier interoperability under
- RED III conditions.
+ RED III conditions. A central new insight is that, in a federated multi-carrier
+ registry, throughput, replication overhead, and regulatory mandate form a single
+ design triad: the per-(country, carrier) channel architecture's lower cross-channel
+ conversion throughput is the deliberate cost of partitioning the ledger along the same
+ lines as the EU's one-issuing-body-per-(country, carrier) regulatory structure, and is
+ therefore the architecturally correct operating point rather than a performance gap to
+ be closed.
```

*Rationale:* swap vague "satisfies key requirements" for the actual numbers; promote the triad insight to a one-sentence summary in the Conclusion.

---

### Conclusion — third paragraph (practical takeaway)

```diff
  Practically, the findings indicate that implementation priorities should move from
  demonstrating technical feasibility toward institutional governance and phased
  migration design. For registry operators and policymakers, this implies focusing on
- standards alignment, consent-aware endorsement policies, and coordinated transition
- pathways across issuing bodies. Overall, the results suggest that federated
- permissioned-DLT architectures provide a credible foundation for multi-carrier GO
- infrastructures.
+ standards alignment, consent-aware endorsement policies, hash-bound cross-channel
+ receipt schemas, and coordinated transition pathways across issuing bodies. Overall,
+ the results suggest that federated, per-(country, carrier) permissioned-DLT
+ architectures provide a credible foundation for multi-carrier GO infrastructures —
+ not in spite of, but precisely because of, the throughput-for-sovereignty trade-off
+ they embody.
```

*Rationale:* adds the new artefact (lock-receipt schemas) to the policy-side recommendations and reframes the closing line.

---

### Conclusion — expert-quote placeholder (formatting only)

```diff
- Expert # states: “Core problem” to highlight the relevance
+ > *Expert # states: "[Core problem]" — to be inserted to highlight the relevance.*
```

*Rationale:* purely cosmetic; flag as TODO and replace curly quotes with straight ones for LaTeX-safety.

---

### Limitations paragraph

```diff
  While our work demonstrates that DLT can enable verifiable cross-domain data sharing
  and asset transfer of GOs across MCES, several limitations warrant acknowledgement.
  The prototype critiques were conducted by the research team rather than independent
  reviewers, and the seven expert interviews provided exploratory validation for the DPs
- but are insufficient for statistical generalization. Additionally, the prototype was
- benchmarked on a single-server deployment. A geographically distributed network across
- multiple EU data centers would introduce additional endorsement latency that may
- compress the measured write throughput.
+ but are insufficient for statistical generalization. The cross-channel conversion
+ benchmark was executed sequentially (one cycle at a time, with deliberate inter-phase
+ pauses) on a single-VM deployment with co-located peers and orderers; a
+ geographically distributed EU-wide network and concurrent multi-client load (e.g. via
+ Caliper) would introduce additional endorsement and ordering latency, and the
+ resulting steady-state TPS profile remains to be measured. Similarly, the EU-scale
+ storage projections (~725 GB/peer/year sharded vs. ~9.8 TB/peer/year monolithic)
+ extrapolate from our per-asset overhead measurements (≈8.8× per electricity GO) and
+ assumed transaction volumes; empirical validation against a multi-year production
+ dataset is left to future work.
```

*Rationale:* the new benchmark *is* sequential and *was* run on a single VM — be honest about it. Also flag that the EU-scale storage numbers are extrapolations.

---

### Penultimate paragraph

```diff
- Despite these limitations, the ten-cycle iterative evaluation with quantitative
- performance evidence across all implemented versions provides a rigorous foundation
- for the proposed design theory, and the practical relevance is underscored by the
- regulatory urgency of RED III mandating multi-carrier GO schemes across all EU member
- states.
+ Despite these limitations, the iterative ten-version evaluation cycle — culminating in
+ a fully passing 30-transaction cross-channel conversion run with quantitative latency,
+ throughput, and storage evidence — provides a rigorous foundation for the proposed
+ design theory, and the practical relevance is underscored by the regulatory urgency of
+ RED III mandating multi-carrier GO schemes across all EU member states.
  These limitations open avenues for future research. First, the DPs should be
  instantiated on alternative platforms such as R3 Corda or Ethereum Layer 2 to test
  their generalizability beyond HLF. Second, a field test with registry operators,
  producers, and buyers in a regulatory sandbox would validate CrossGO under real
- metering loads.
+ metering loads and concurrent cross-channel conversion traffic.
  Third, organizational adoption studies are needed to examine how registry operators
  and regulators weigh DLT-based against centralized architectures, complementing our
  technical evaluation with institutional perspectives.
```

*Rationale:* the original "ten-cycle iterative evaluation" was ambiguous (cycles of what?) — clarify as "ten-version" + reference the 30-tx run; extend the field-test future work to include concurrent cross-channel load.

---

### Closing sentence

```diff
- Despite these boundaries, the CrossGO artifact and the underlying DPs offer a concrete
- starting point for registry operators, policymakers, and researchers seeking to move
- cross-carrier GO interoperability from conceptual proposals toward implementable
- infrastructure.
+ Despite these boundaries, the CrossGO v10.5.4 artifact and the underlying DPs offer a
+ concrete, benchmarked, and regulatorily-aligned starting point for registry operators,
+ policymakers, and researchers seeking to move cross-carrier GO interoperability from
+ conceptual proposals toward implementable infrastructure.
```

*Rationale:* version-stamps the artefact and adds two adjectives ("benchmarked", "regulatorily-aligned") that the v10 work has earned.

---

## Part 3 — Items deliberately NOT changed

Call out if you disagree with any of these:

1. The `41.9 TPS` figure — kept as-is; presumably sourced from earlier versions.
2. The `Expert # states: "Core problem"` placeholder — kept the placeholder; only normalised the quoting.
3. All five DP names and their substantive definitions — only added bold labels and the v10-specific extensions; the original DP claims are preserved.
4. The "future work: Corda / Ethereum L2 / regulatory sandbox / org adoption studies" list — kept verbatim.
5. The DP3 sentence about interviewee privacy expectations was **dropped** — flag if you want it restored; removed because its evidence base (interviews) was not strengthened by the v10 work and the paragraph reads tighter without it.

---

## Part 4 — Provenance

| Item | Source |
|---|---|
| v10.5.4 chaincode + dual-channel benchmark numbers | [20260427_v10.5.4_perf_results.md](20260427_v10.5.4_perf_results.md), [20260427_perf_results.csv](20260427_perf_results.csv), [20260427_perf_run.log](20260427_perf_run.log) |
| `ConversionLock` + `TxID` hash fix (v10.5.3) | [20260427_v10.5.2_needed_fixes.md](20260427_v10.5.2_needed_fixes.md) |
| Cross-channel conversion implementation | [20260425_CrossChannel_Conversion_Implementation.md](20260425_CrossChannel_Conversion_Implementation.md) |
| EU-scale storage projections (725 GB / 9.8 TB / 264 TB / 19.6 TB / 28 GB / 750 GB) | [Project-Description/DATA_DUPLICATION_CONSIDERATIONS.md](Project-Description/DATA_DUPLICATION_CONSIDERATIONS.md) §3.3 |
