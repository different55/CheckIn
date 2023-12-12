# CheckIn

CheckIn is a way to let other people know what the heck you're up to.

It's intended for multi-user systems like tilde.town.

It's essentially a status. But like. For your terminal. Has that been done
already? Has to have been, right?

## Usage:

`checkin set [--include-wd]`

`--include-wd` will include your current working directory in your message.

`checkin get [--freshness=14]`

`--freshness` controls how new (in days) messages must be in order to be printed.

## Example:


```
diff:~/p/Apricitas > checkin set --include-wd "is trying to wrap his lil head around making a raytracer" 
diff:~/p/Apricitas > checkin get
   ~diff is trying to wrap his lil head around making a raytracer. (~diff/projects/Apricitas)
   ~exampleuser is busy setting an example.
```