+++
slug = "mathematica-multiplicative-models"
date = 2020-11-05
visibility = "published"
bib_paths = ["/ref.bib"]
+++

# Multiplicative models for $2^2r$ experiments

Chapter 18 of *The Art of Computer Systems Performance Analysis* [^@raj1991art] 
covers multiplicative models for $2^2r$ experiments. The additive model for
analysis of a $2^2r$ experiment was assumed:

$$
y_{ij} = q_0 + q_A x_A + q_B x_B + q_{AB} x_A x_B + e_{ij}
$$

The additive model assumes the effect of the factors, their interactions, and 
the errors are additive. This assumption doesn't hold for some workloads. Jain
provides the example of measuring the performance of processors on different 
workloads.

> Suppose the measure response $y_ij$ represents the time required to execute
> a workload of $w_j$ instructions on a processor capable of executing $v_i$
> instructions per second. Then if there are no errors or interactions, we know 
> that the time would be $y_{ij} = v_i w_j$. The effects of the two factors are
> not additive; they are multiplicative.

The convert a multiplicative model to an additive model, we use a log transform:

$$
log(y_{ij}) = log(v_i) +  log(w_j)
$$

Then, we can use a modified additive model.

$$
y'_{ij} = q_0 + q_A x_A + q_B x_B + q_{AB} x_A x_B + e_{ij}
$$

Where $y'_{ij} = log(y_{ij})$ represents the transformed response. Similarly,
we can apply the antilog to the effects $q_A$, $q_B$, and $q_{AB}$ to produce
multiplicative effects $u_A = 10^{q_A}$, $u_B = 10^{q_B}$, and 
$u_{AB} = 10^{q_{AB}}$.

> The $u_A$ so obtained would represent the ratio of the MIPS rating of the two
> processors. Similarly, $u_B$ represents the ratio of the size of the two 
> workloads. The antilog of additive mean $q_0$ produces the geometric mean of 
> the responses:
>
> $$ \dot y = 10^{q_0} = (y_1 y_2 \cdots y_n)^{1/n} \;\;\;\;\; n = 2^2r $$



