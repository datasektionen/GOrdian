# Budgetsystemet GOrdian

Fullstack Golang application that handles the budget of Datasektionen.

Live at [budget.datasektionen.se](https://budget.datasektionen.se).

# API

**API documentation** can be found here: **TBA**

# Local
dunno lol

# Entities

Each entry in the system consists of the following entities.
There are three elementary entities in the 
system, Cost centres, Secondary cost centres and Budget lines.

## Cost Centre
A *Cost Centre* is called *Resultatställe* in Swedish. It can be typed as a Project (Projekt), Committee (Nämnd) or Other (Övrigt).
They are the main budget divisions.

## Secondary cost centre
A *Secondary cost centre* is part of a Cost centre. It can be an event or otherwise differentiable part of the Cost centre.
Each Secondary cost centre contains the Cost centre it is connected to.

## Budget line
A *Budget line* is the smallest part of a budget, belonging to a Secondary cost centre.
Each Budget line centre contains the Secondary cost centre it is connected to.
A Budget Line also contains the following.

### Income and Expense
The prediction of what someone expects to earn or spend.
A Budget Line never has an Income and an Expense at the same time.

### Account
*Account* is used to assist with the bookkeeping

### Comment
A *Comment* is used to further clarify the use of a Budget line

## Very nice money budget