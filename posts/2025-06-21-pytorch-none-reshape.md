---
title: "PyTorch Reshaping with None"
date: 2025-06-21
author: detorch
---

## PyTorch Reshaping with None 

Currently I am learning attention mechanism from Dive into Deep Learning book. In the book I see following implementation in masked softmax:

```python
def sequence_mask(X, valid_len, value= -1e6):
    """ X is 2D array (number_of_points, maxlen), valid_len is 1D array (number_of_points)"""
    max_len = X.size(1)

    mask = torch.arange(max_len, dtype=torch.float32, device=X.device)[None, :] < valid_len[:, None]
    X[~mask] = value
    return X
```

In sequential data processing, I mean processing natural language. The sequence length might be variable for each data point. For example : 

1 : "Welcome To My Blog"


2 : "Hello World"

To solve that problem , we fill remaining values with a special token. 

1 : "Welcome To My Blog"


2 : "Hello World blnk blnk"


In attention, we do not want to attend to blnk tokens. So we create mask for that. In the code portion max_len is the maximum length of the sequence and valid_len is the actual length of the sequence. I mean for 1st data point valid_len is 3 and for 2nd data point valid_len is 2.

In the code portion, we are trying to create mask for that. Let's say we have following dictionary ['blnk', 'Welcome', 'To', 'My', 'Blog', 'Hello', 'World'] so X vector will be : 
```python
X = [
    [1,2,3,4],
    [5,6,0,0]
]
valid_len = [3,2]
```

and the mask must be : 
```python
[
    [1,1,1,0],
    [1,1,0,0]
]
```

We are using brodcast mechanism to create mask. First `torch.arange(max_len, dtype=torch.float32, device=X.device)` will create a 1D array with shape (max_len,). In our example, it would be [0,1,2,3]. It must be [[0,1,2,3],[0,1,2,3]] right? But we will use brodcast mechanism to expand it to [[0,1,2,3],[0,1,2,3]]. For getting our broadcasted mask we need to apply operator to lengths of (max_len, 1) and (1, valid_len). If you do not know broadcast mechanism, you can read about it in [PyTorch documentation](https://pytorch.org/docs/stable/notes/broadcasting.html). 

Now we came to point, for reshaping we use `None` in pytorch. `torch.arange(max_len, dtype=torch.float32, device=X.device)[None, :]` is equivalent to `torch.arange(max_len, dtype=torch.float32, device=X.device).reshape(1, -1)` and  `valid_len[:, None]` is equivalent to `valid_len.reshape(-1, 1)`. 

To be honest, I would prefer reshape more readable so my version of this function is : 
```python
def sequence_mask_with_reshape(X, valid_len, value= -1e6):
    """ X is 2D array (number_of_points, maxlen), valid_len is 1D array (number_of_points)"""
    max_len = X.size(1)
    mask = torch.arange(max_len, dtype=torch.float32, device=X.device).reshape(1, -1) < valid_len.reshape(-1, 1)
    X[~mask] = value
    return X
```

As you can see, it is more readable.