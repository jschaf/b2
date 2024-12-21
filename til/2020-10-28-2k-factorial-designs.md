+++
slug = "2k-factorial-designs"
date = 2020-10-28
visibility = "published"
+++

# $2^k$ factorial designs in Mathematica

Chapter 17 of *The Art of Computer Systems Performance Analysis* covers $2^k$
factorial designs. A $2^k$ factorial design determines the effect of $k$ factors
where each factor has two *levels* or alternatives. The $2^k$ factorial design
is useful at the start of a performance study to reduce the amount of detail
needed for a full factorial design. Most factors are unidirectional, so a $2^k$
factorial design only examines the min and max values for each factor. If the
range of performance for a factor is small, you can likely stop with just 2 
factors saving time and cost for a full factorial design.

## $2^2$ factorial design

We'll use the example from the book. The performance in million instructions per
second (MIPS) is measured by varying the cache size of 1 and 2 KB, and the 
memory size of 4 MB and 16 MB.

| Cache size (KB) | MIPS (Memory size = 4 MB) | MIPS (Memory Size = 16 MB) |
|----------------:|--------------------------:|---------------------------:|
|               1 |                        15 |                         45 |
|               2 |                        25 |                         75 |

We'll define two variables: $x_A$ is -1 for 4MB of memory and 1 for 16 MB of 
memory. $x_B$ is -1 for 1 KB of cache and 1 for 2 KB of cache.

The performance $y$ in MIPS is regressed on $x_A$ and $x_B$ with a nonlinear 
regression:

$$
y = q_0 + q_A x_A + q_B x_B + q_{AB} x_A x_B
$$

Substituting the 1 and -1 for $q_A$ and $q_B$ as well as the observation $y$ 
yields:

$$
15 = q_0 - q_A - q_B + q_{AB} \\
45 = q_0 + q_A - q_B - q_{AB} \\
25 = q_0 - q_A + q_B - q_{AB} \\
75 = q_0 + q_A + q_B + q_{AB} \\
$$

Solving the regression equation for $y$ is:

$$
y = 40 + 20x_A + 10 x_B + 5 x_A x_B
$$

In general, solving for $q_i$'s by using the four observations $y_i$:

$$
y_1 = q_0 - q_A - q_B + q_{AB} \\ 
y_2 = q_0 + q_A - q_B - q_{AB} \\ 
y_3 = q_0 - q_A + q_B - q_{AB} \\ 
y_4 = q_0 + q_A + q_B + q_{AB} \\ 
$$

$$
\begin{aligned}
q_0    &= \tfrac{1}{4}( y_1 + y_2 + y_3 + y_4) \\  
q_A    &= \tfrac{1}{4}(-y_1 + y_2 - y_3 + y_4) \\ 
q_B    &= \tfrac{1}{4}(-y_1 - y_2 + y_3 + y_4) \\ 
q_{AB} &= \tfrac{1}{4}( y_1 - y_2 - y_3 + y_4) \\ 
\end{aligned}
$$

### Allocation of variation

The sample variance of 
$y = s_y^2 = \frac{\sum_{i=1}^{2^2} (y_i - \bar y)^2 }{ 2^2 - 1}$.
The numerator is of the faction is the total variation of $y$, or SST. The 
variation consists of 3 parts:

$$
\begin{aligned}
SST = 2^2 q^2_A& + 2^2 q_B^2& + 2^2 q_{AB}^2& \\
SST = SSA& + SSB& + SSAB& \\
\end{aligned}
$$






