# memorydb
In memory database with nested transactional support

# Intro
A very popular interview question asked in coding interviews and this is my implementation of memory db.
Different versions of interviews ask different versions of this question, but the basic gist is the same.
Depending on the available time - (45 mins, or 1 hour 15 mins, or 2 hours), the complexity of this question
and expectations asked changes and we'll cover all. This implemetation covers from basic CRUD operations of a key-value to handling
complex transactions like nested transactions.

# Demo
![](demo_memorydb.gif)

```
&{data      count       deletedKey  deletedValue    root    level   parent}
&{map[A:10] map[10:1]   map[A:true] map[10:1]       true    0       <nil>}

```
# Problem Statement(s)

1) Part 1: Basic CRUD Operations \
(This is usually asked as a Part 1 of the problem) \
Q: Design and implement an in-memory key-value datastore. It should support basic operations - 
`GET`, `SET`, `DELETE` of string keys and values.
Another complexity can be added by asking to implement a `COUNT` function which gets you the count of a value

Eg:
```
db.set("A", 100)
print(db.get("A")) // returns 100
db.count(a) // returns 1
db.delete("A")
db.get("A") // error or NULL
```

2) Part 2: Implementent Transactions\
(This is usually asked as a Part 2 of the problem once you finish Part 1)\
Q: A transaction is created with the `BEGIN` command and creates a context for the other operations to happen. Until the active transaction is committed using the `COMMIT` command, those operations do not persist. And, the `ROLLBACK` command throws away any changes made by those operations in the context of the active transaction.

Eg 1:
```
db.set("A", 100)
print(db.get("A"))
db.begin()
db.get("A") // returns 100 - set before transaction began
db.set("A", 200) // returns 200
db.commit()
db.get("A") // returns 200 
```

Eg 2:
```
db.set("A", 100)
print(db.get("A"))
db.begin()
db.get("A") // returns 100 - set before transaction began
db.set("A", 200)
db.delete("A") // key A is deleted in current context
db.get("A") // returns ERROR or NULL since A was deleted
db.rollback() // throws away the whole transactions
db.get("A") // returns 100 since it was originally set as 100
```
As you saw, what happens in transaction, stays in a transaction. It doesn't 
persists and can be rolled back to the original state of database.

2) Part 3: Implementent Nested Transactions \ 
(This is sometimes asked with Part 2 of the problem or can be asked as a Part 3. However, due to it's complexity, usually asked in long coding interviews i.e. more than 1 hour sessions)\
Q: The memory database should support nested transaction. You begin a transaction with a `BEGIN` command. However, another `BEGIN` command within the previous begin will start a new transaction. `ROLLBACK` will only rollback the most recent transaction. `COMMIT` will apply all transactions together.

Eg 1:
```
db.begin()
db.set("A", 100)
db.get("A") // returns 100

db.begin()
db.set("A", 200)
db.get("A") // returns 200

db.rollback()
db.get("A") // returns 100

db.rollback()
db.get("A") // ERROR or NULL

db.commit() // ERROR - Nothing to commit since both transactions were rolled back
```


Eg 2:
```
db.begin()
db.set("A", 100)
db.get("A") // returns 100

db.begin()
db.set("A", 200)
db.get("A") // returns 200

db.delete("A")
db.get("A") // ERROR or NULL

db.commit()
db.get("A") // ERROR or NULL
```

Eg 3:
```
db.begin()
db.set("A", 100)
db.get("A") // returns 100

db.begin()
db.set("A", 200)
db.get("A") // returns 200

db.delete("A")
db.get("A") // ERROR or NULL

db.commit()
db.get("A") // ERROR or NULL
```

# Thought Process and intuition

Obviously, you need a map (dict) to save the key and value for quick lookups for db.get()
You need another map (dict) to save value and the number of times that value has occurred.

As you have noticed, whenever a new transaction starts, a new state of database is started and whatever happens in that state remains in that state. If you delete a key, the key is only deleted in that transaction. If you re-set a key with a new value, the key is only set in that trasnaction. Unless you commit. How will you handle that?

Well, one way is to start a new instance of your db object, save the parent's pointer, copy over value of "A" when you begin a transaction. That way, when you do db.get("A"), you still get the value which was set before the current transaction began. And if you delete "A", it'll only delete from this current instance. This was actually an acceptable solution in one of the interviews - however, it's not the most efficient one. Every BEGIN command doubles the memory usage as you need to copy over all the keys.

[PR#2 ](https://github.com/ranesagar/memorydb/pull/2) optimizes this. Two additonal maps can be introducted with saves the deleted values i.e.  `deletedKey` and `deletedValue` maps keep track of deleted keys and corresponding occurrences of upserted keys in the current transaction.

get() - Needs to change. It should now recursively searches for the key till root node
set() - Before set a value, check with get() if the value exists and then handle the upsert.
commit() - Deep merges each map as we go up till the root node.

Enjoy the code and feel free to suggest any optimizations :) 
