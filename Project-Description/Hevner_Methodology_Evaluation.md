# Evaluation of ICIS 2026 Paper Against Hevner et al. (2004) DSR Framework

**Paper evaluated:** "All Carriers Allowed: Design Principles for Distributed Ledger-Based Cross-Domain Data Sharing Between Guarantee of Origin Schemes"

**Reference framework:** Hevner, A. R., March, S. T., Park, J., & Ram, S. (2004). Design Science in Information Systems Research. *MIS Quarterly*, 28(1), 75–105.

**Supplementary:** Hevner, A. R. (2007). A Three Cycle View of Design Science Research. *Scandinavian Journal of Information Systems*, 19(2), 87–92.

---

## 1. Overview

This document evaluates whether the ICIS 2026 DSR paper soundly follows the Hevner et al. (2004) framework. The evaluation is structured around (a) the seven DSR guidelines from Table 1 of Hevner et al. (2004), and (b) the three-cycle view from Hevner (2007). For each guideline, a verdict of **Met**, **Partially Met**, or **Not Met** is given, along with supporting evidence and recommendations.

---

## 2. Evaluation Against the Seven DSR Guidelines

### Guideline 1: Design as an Artifact

> *"Design-science research must produce a viable artifact in the form of a construct, a model, a method, or an instantiation."*

**Verdict: Met**

| Criterion | Evidence |
|-----------|----------|
| Artifact type | Instantiation (HLF prototype) + Constructs (four design principles) |
| Viability | v7.0 prototype: 10 contract namespaces, ~50 exported functions, deployed on production-grade infrastructure (Hetzner 16 vCPU, 32 GB RAM) |
| Beyond trivial | 29 ADRs across 8 design cycles; v8.0 multi-channel architecture design extends beyond incremental feature addition |

**Assessment:** The paper produces two distinct artifact types—an IT artifact (the Hyperledger Fabric prototype) and design knowledge artifacts (four design principles). The prototype is not a proof-of-concept toy but a production-grade platform with quantitative performance evidence. The v8.0 architecture design adds a further artifact (architectural model) that addresses structural limitations of the instantiation. This exceeds the minimum threshold of Guideline 1.

---

### Guideline 2: Problem Relevance

> *"The objective of design-science research is to develop technology-based solutions to important and relevant business problems."*

**Verdict: Met**

| Criterion | Evidence |
|-----------|----------|
| Business problem | Cross-carrier GO interoperability mandated by RED III directive; 700M+ certificates/year market |
| Stakeholder relevance | Issuing bodies (27 EU national authorities), energy producers, consumers, regulators |
| Regulatory urgency | RED III mandates GO schemes for hydrogen, biogas, district heating by 2024 |
| Gap identification | 8 challenges (C1–C8) systematically identified via multi-vocal literature review from 44 sources |

**Assessment:** The paper makes a compelling case for problem relevance. The regulatory mandate (RED III) creates immediate urgency, and the 700M+ certificate market demonstrates scale significance. The conversion issuance problem is well-articulated as a design challenge requiring cross-domain data sharing between previously separate registries. Section 1.2 explicitly connects C1–C8 to documented industry pain points with specific literature citations.

**Recommendation:** A brief quantitative estimate of the economic cost of the status quo (e.g., transaction cost overhead from manual auditing, estimated losses from double counting) would further strengthen the relevance argument.

---

### Guideline 3: Design Evaluation

> *"The utility, quality, and efficacy of a design artifact must be rigorously demonstrated via well-executed evaluation methods."*

**Verdict: Met (strong for v7.0; partial for v8.0)**

| Evaluation Method | Application in Paper |
|-------------------|---------------------|
| **Observational** | Not used (no field study or case study with real GO registry operators) |
| **Analytical** | Architecture critiques after each cycle (arguments in favour / against per DP) |
| **Experimental** | Caliper benchmarks: 28 rounds, 511 seconds, 10 workers, 901 devices (v7.0) |
| **Testing** | Functional testing implied by 100% success rates; regression testing across v3–v7 |
| **Descriptive** | Detailed description of artifact (Sections 4.4–4.6), design principles (Table 3) |

**Assessment:** The evaluation is rigorous for v7.0. The concurrent evaluation approach (Sonnenberg and vom Brocke 2012) combining quantitative Caliper benchmarks with systematic architecture critiques is well-suited to the artifact type. The multi-version benchmarking (v3.0, v5.0, v7.0) demonstrates stability and non-regression—a form of empirical evidence rarely seen in DSR papers. The DP-Challenge-MR mapping table (Table 4) provides clear traceability from problem to solution evidence.

**Limitation:** The v8.0 multi-channel architecture is evaluated only qualitatively. The paper correctly notes this limitation and distinguishes between "confirmed" properties (from Fabric documentation) and "projected" properties (requiring benchmarking). This is acceptable for a designed but not-yet-implemented cycle, but the paper should be transparent that the v8.0 evaluation carries less evidentiary weight than v1.0–v7.0.

**Recommendation:** Consider adding a brief note in Section 5.1 that v8.0 undergoes qualitative evaluation only, and frame the v8.0 discussion (Section 6.2) explicitly as a "projected evaluation" to manage reader expectations about evidence strength.

---

### Guideline 4: Research Contributions

> *"Effective design-science research must provide clear and verifiable contributions in the areas of the design artifact, design foundations, and/or design methodologies."*

**Verdict: Met**

| Contribution Area | Evidence |
|-------------------|----------|
| **Design artifact** | HLF prototype (v1.0–v7.0 implemented, v8.0 designed); 10 namespaces, ~50 functions |
| **Design foundations** | Four DPs formulated via Gregor et al. (2020) anatomy; positioned as Level 2 nascent design theory (Gregor & Hevner 2013) |
| **Design methodology** | ADR-based iterative cycle methodology; concurrent evaluation approach |

**Assessment:** The paper provides contributions in all three areas identified by Hevner et al. The design principles are articulated using a recognized schema (Gregor et al. 2020), making them separable from the specific artifact and applicable to other contexts (e.g., DP1's hash-based ID generation for any HLF asset system). The ADR-based methodology is a genuine methodological contribution that demonstrates how software engineering practices can be integrated into DSR. The paper could more explicitly articulate the *novelty* of each contribution (i.e., what is known vs. new).

**Recommendation:** Strengthen the novelty claim by explicitly comparing each DP against the closest prior design knowledge. For example, DP1's hash-based ID mechanism is seemingly simple but demonstrates a non-obvious interaction between application-layer design and platform-level MVCC—make this novelty explicit rather than relying on the reader to infer it.

---

### Guideline 5: Research Rigor

> *"Design-science research relies upon the application of rigorous methods in both the construction and evaluation of the design artifact."*

**Verdict: Met**

| Aspect | Evidence of Rigor |
|--------|-------------------|
| **Construction rigor** | Systematic literature review (32 academic + 12 grey sources, 4 databases); Gioia et al. (2013) coding methodology for challenge identification; blockchain suitability assessment via Gramlich et al. (2023) |
| **Evaluation rigor** | Caliper v0.6.0 benchmarks with specified parameters (10 workers, fixed TPS send rates); multiple versions benchmarked (v3.0, v5.0, v7.0); architecture critique methodology with balanced "arguments in favour / against" |
| **Formalization** | Gregor et al. (2020) DP anatomy; Peffers et al. (2007) DSR methodology; Sonnenberg and vom Brocke (2012) concurrent evaluation |
| **Traceability** | C1–C8 → MR1–MR4 → DP1–DP4 → Instantiation evidence; 29 ADRs providing decision audit trail |

**Assessment:** The paper demonstrates strong methodological rigor. The challenge identification follows established qualitative coding methods (Gioia et al. 2013), the DP formulation uses a validated schema (Gregor et al. 2020), and the evaluation employs both quantitative (benchmarks) and qualitative (critiques) methods. The ADR-based documentation provides an auditable decision trail that is commendable for traceability. The multi-version benchmarking across v3.0–v7.0 provides a rare longitudinal dimension to the evaluation.

**Minor concern:** The paper does not report inter-rater reliability for the challenge coding (C1–C8), and the architecture critiques appear to be conducted by the research team rather than independent experts. This is common in single-author DSR but should be acknowledged as a validity limitation.

---

### Guideline 6: Design as a Search Process

> *"The search for an effective artifact requires utilizing available means to reach desired ends while satisfying laws in the problem environment."*

**Verdict: Met (strongly)**

| Criterion | Evidence |
|-----------|----------|
| Iterative search | 8 design cycles (v1.0–v8.0) with documented iteration rationale |
| Design space exploration | 29 ADRs documenting alternatives considered and reasons for chosen approaches |
| Problem-solution fitness | Each cycle addresses gaps identified in previous cycle's critique |
| Means-ends analysis | Explicit mapping from challenges → meta-requirements → design principles → instantiation |

**Assessment:** This is perhaps the paper's strongest alignment with Hevner et al. The eight-cycle iteration process with 29 documented ADRs is an exemplary demonstration of design as a search process. Each ADR documents the problem encountered, alternatives considered, chosen approach, and consequences—directly mapping to Hevner et al.'s (2004) concept of searching for effective artifacts using available means. The progression from 0.8 TPS (v1.0) to 50 TPS (v3.0) through a single design change (ADR-001) illustrates the power of informed search in the design space. The v8.0 cycle demonstrates continued search beyond local optima, recognizing that the single-channel architecture—despite incremental improvements across v4.0–v7.0—could not resolve the fundamental scalability ceiling.

---

### Guideline 7: Communication of Research

> *"Design-science research must be presented effectively both to technology-oriented as well as management-oriented audiences."*

**Verdict: Met**

| Audience | Evidence |
|----------|----------|
| **Technology-oriented** | Go chaincode details, Fabric configuration, Caliper benchmark parameters, CouchDB specifics, ECDSA P-256 implementation, state-based endorsement policies |
| **Management-oriented** | Section 6.4 practical implications (actionable insights for registry operators and regulators); DP-level abstractions above platform specifics; cost analysis (8.8× storage overhead) |
| **Academic** | Standard ICIS paper structure; DPs formulated per Gregor et al. (2020); positioned as nascent design theory per Gregor & Hevner (2013) |

**Assessment:** The paper targets the ICIS venue appropriately. The abstract and introduction are accessible to a management audience, while the prototype details (Sections 4.4–4.6) serve a technical audience. The discussion successfully bridges both through the four practical implications. The use of the Gregor et al. (2020) DP anatomy ensures that the prescriptive contributions are stated in a form familiar to the IS design science community.

**Recommendation:** The paper is technically dense. For the ICIS proceedings format, consider whether some implementation details (e.g., specific CouchDB configurations, Go struct field names) can be moved to an online appendix to improve readability for the management-oriented subset of the ICIS audience.

---

## 3. Evaluation Against the Three-Cycle View (Hevner 2007)

### 3.1 Relevance Cycle

> *Connects the research to the environment: people, organizational systems, technical systems, problems, and opportunities.*

| Element | Evidence |
|---------|----------|
| **People** | Issuing bodies, producers, consumers, regulators (Section 2.1) |
| **Organizational systems** | National GO registries, AIB Hub, CEN standardization body |
| **Technical systems** | Centralized registry IT systems, EECS message protocols, metering devices |
| **Problems** | 8 challenges (C1–C8) from multi-vocal literature review |
| **Opportunities** | RED III mandate, CEN-EN 16325 revision enabling multi-carrier GOs |

**Verdict: Well-established.** The relevance cycle inputs (C1–C8) are systematically derived from 44 literature sources. The outputs (DPs and prototype) are framed with explicit practical implications (Section 6.4). The field testing gap (no deployment with real GO registry operators) is acknowledged in limitations.

**Recommendation:** Strengthen the relevance cycle feedback by describing how findings have been (or could be) communicated back to GO market stakeholders (e.g., via the CEN/CLC/JTC 14 standardization committee, AIB working groups, or regulatory consultations).

---

### 3.2 Rigor Cycle

> *Connects the research to the knowledge base: scientific theories, experience, expertise, meta-artifacts (design products and design processes), and existing artifacts.*

| Element | Evidence |
|---------|----------|
| **Theories** | DLT theory (Androulaki et al. 2018), blockchain suitability frameworks (Gramlich et al. 2023; Wüst & Gervais 2018), DSR methodology (Peffers et al. 2007) |
| **Experience & expertise** | Master thesis (Albrecht 2024) as Phase 1; 7 subsequent cycles building on accumulated domain knowledge |
| **Meta-artifacts** | Gregor et al. (2020) DP anatomy schema; Gioia et al. (2013) coding methodology; Sonnenberg & vom Brocke (2012) concurrent evaluation |
| **Existing artifacts** | AIB Hub EECS transfer protocol; CEN-EN 16325 standard; ENTSO-E Transparency Platform; prior HLF prototypes (Chuang et al. 2019; Ølnes et al. 2022) |

**Verdict: Well-grounded.** The paper draws on a broad knowledge base spanning DLT, energy informatics, IS design theory, and software engineering. The cumulative learning across 8 cycles demonstrates experience-based design refinement. The ADR methodology represents an integration of software engineering practice into DSR.

**Recommendation:** More explicitly connect the hash-based ID mechanism (ADR-001, DP1) to the broader literature on MVCC performance in distributed databases—this connects the construction knowledge to database theory rather than presenting it as an ad hoc discovery.

---

### 3.3 Design Cycle

> *The tighter loop of building and evaluating the design artifact and processes.*

| Phase | Build Activity | Evaluate Activity |
|-------|---------------|-------------------|
| v1.0–v3.0 | Conceptual design, initial prototype, refined prototype | Literature-based evaluation, initial benchmarks |
| v4.0–v5.0 | Pagination, tombstones, commitments, CEN alignment, biogas | Caliper v0.6.0 benchmarks, v5.0 architecture critique |
| v6.0–v7.0 | Security hardening, bridge, device attestation, oracle | v7.0 benchmarks (28 rounds, 511s, 901 devices), v7.0 critique |
| v8.0 | Multi-channel architecture design, cross-channel bridge protocol | Qualitative DP evaluation (pending implementation) |

**Verdict: Exemplary.** The design cycle is the paper's strongest methodological dimension. Eight iterations with documented build-evaluate cycles—each informed by gap identification from the previous cycle's critique—demonstrate the tight design loop Hevner (2007) prescribes. The traceable chain from critique → ADR → implementation → benchmark → critique is a model for DSR artifact development.

**Minor concern:** The v8.0 cycle intentionally breaks the build-evaluate pattern by presenting only the build (design) phase without the evaluate (benchmark) phase. This is acknowledged in the paper, but it introduces an asymmetry in the evidence base. Framing v8.0 as a "design-ahead" rather than a completed cycle would be more methodologically precise.

---

## 4. Summary Assessment

| Guideline | Verdict | Strength |
|-----------|---------|----------|
| G1: Design as an Artifact | **Met** | Strong — dual artifact (prototype + DPs) |
| G2: Problem Relevance | **Met** | Strong — regulatory mandate + 700M market |
| G3: Design Evaluation | **Met** | Strong for v7.0; partial for v8.0 |
| G4: Research Contributions | **Met** | Strong — all three contribution areas |
| G5: Research Rigor | **Met** | Strong — multi-method, multi-version |
| G6: Design as a Search Process | **Met** | Exemplary — 8 cycles, 29 ADRs |
| G7: Communication of Research | **Met** | Adequate — dual audience addressed |

| Cycle | Verdict | Strength |
|-------|---------|----------|
| Relevance Cycle | **Well-established** | Systematic problem identification; practical implications |
| Rigor Cycle | **Well-grounded** | Broad knowledge base; validated schemas |
| Design Cycle | **Exemplary** | 8 iterations, traceable build-evaluate chain |

---

## 5. Overall Verdict

**The paper soundly follows the Hevner et al. (2004) DSR framework.** All seven guidelines are met, and the three-cycle view is well-articulated. The iterative design process (Guideline 6) and the concurrent evaluation methodology (Guideline 3) are particular strengths that exceed the typical standard for DSR publications at top IS venues. The addition of v8.0 introduces a slight methodological asymmetry (design without implementation/benchmarking), but this is transparently acknowledged and does not undermine the overall rigor of the contribution.

---

## 6. Recommendations for Strengthening Methodological Soundness

1. **Frame v8.0 explicitly as an incomplete design cycle.** Add a sentence in Section 3.2 (Iterative Software Design) noting that the eighth cycle presents a validated design awaiting implementation and benchmarking, distinguishing it from the seven completed build-evaluate cycles.

2. **Acknowledge single-researcher evaluation validity.** Note in Section 5.1 that architecture critiques were conducted by the research team rather than independent domain experts, and that inter-rater reliability was not assessed for the challenge coding (C1–C8).

3. **Strengthen the relevance cycle feedback loop.** Describe any engagement with GO market practitioners (AIB, CEN/CLC/JTC 14, national issuing bodies) that informed the design or validated the results. Even informal validation (e.g., expert review of DPs) would strengthen the relevance cycle.

4. **Consider a structured evaluation matrix.** Adopt one of Hevner et al.'s (2004) evaluation method categories (Table 2) as a formal structure: explicitly state which methods were applied (experimental, analytical, descriptive) and which were not (observational, testing in real environments), with justification for the choices.

5. **Quantify the economic relevance.** Add a brief calculation of the cost savings or efficiency gains the architecture could provide relative to the status quo (manual auditing costs, double-counting losses, platform operator fees) to make the practical contribution more tangible for management-oriented readers.

6. **Connect DP1 to database theory.** Explicitly reference the MVCC literature (e.g., Bernstein & Goodman 1981 on concurrency control in distributed databases) when discussing the contention-free scalability mechanism, grounding the construction decision in the rigor cycle's knowledge base.
