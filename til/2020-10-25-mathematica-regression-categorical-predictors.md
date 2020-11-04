+++
slug = "mathematica-regression-categorical-predictors"
date = 2020-10-25
visibility = "published"
bib_paths = ["/ref.bib"]
+++

# Linear regression with categorical predictors in Mathematica

Chapter 15 of *The Art of Computer Systems Performance Analysis* [^@raj1991art] 
covers linear regression with categorical predictors.

If all the variables are categorical, Jain recommends a factorial design 
instead. Jain also notes that the factorial designs will yield more precise 
We'll use the sample data for results with less variation.


TABLE: Measured RPC times on Unix and Argus

| System | Data size (bytes) | Time (ms) |
|--------|------------------:|-----------:|
| unix   | 64                | 26.4       |
| unix   | 64                | 26.4       |
| unix   | 64                | 26.4       |
| unix   | 64                | 26.2       |
| unix   | 234               | 33.8       |
| unix   | 590               | 41.6       |
| unix   | 846               | 50.0       |
| unix   | 1060              | 48.4       |
| unix   | 1082              | 49.0       |
| unix   | 1088              | 42.0       |
| unix   | 1088              | 41.8       |
| unix   | 1088              | 41.8       |
| unix   | 1088              | 42.0       |
| argus  | 92                | 32.8       |
| argus  | 92                | 34.2       |
| argus  | 92                | 32.4       |
| argus  | 92                | 34.4       |
| argus  | 348               | 41.4       |
| argus  | 604               | 51.2       |
| argus  | 860               | 76.0       |
| argus  | 1074              | 80.8       |
| argus  | 1074              | 79.8       |
| argus  | 1088              | 58.6       |
| argus  | 1088              | 57.6       |
| argus  | 1088              | 59.8       |
| argus  | 1088              | 57.4       |


First, we'll input into mathematica:

```mathematica
data = {
  {"unix", 64, 26.4},
  {"unix", 64, 26.4},
  {"unix", 64, 26.4},
  {"unix", 64, 26.2},
  {"unix", 234, 33.8},
  {"unix", 590, 41.6},
  {"unix", 846, 50.},
  {"unix", 1060, 48.4},
  {"unix", 1082, 49.},
  {"unix", 1088, 42.},
  {"unix", 1088, 41.8},
  {"unix", 1088, 41.8},
  {"unix", 1088, 42.},
  {"argus", 92, 32.8},
  {"argus", 92, 34.2},
  {"argus", 92, 32.4},
  {"argus", 92, 34.4},
  {"argus", 348, 41.4},
  {"argus", 604, 51.2},
  {"argus", 860, 76.},
  {"argus", 1074, 80.8},
  {"argus", 1074, 79.8},
  {"argus", 1088, 58.6},
  {"argus", 1088, 57.6},
  {"argus", 1088, 59.8},
  {"argus", 1088, 57.4}
}
```

Computing the linear model only requires declaring nominal variables:

```mathematica
lm = LinearModelFit[data, {type, bytes}, {type, bytes}, NominalVariables -> type]

(* Fitted Model [21.8124 + 0.0252066 bytes + 14.9266 DiscreteIndicator[type, argus, {argus, unix}] *)
```

The model can be interpreted as follows. The setup cost for both Unix and Argus 
is 21.8124 ms. Argus has an additional setup cost of 14.9266 ms. The per-byte
processing time is 0.025 ms.

Jain ends the section by noting that this model is only valid if both the Unix
and Argus systems use the same code path. Otherwise, two separate, simple linear 
regression models, one for Unix and one for Argus would be more realistic.
