This is a CLI application where user prompts for an object:

- a cat
- a box
- a house
- a cake
or whatever the user types in


And it executes (inside k/ folder)

./run -m k-1b-q8_0.gguf -p "Hey cadmonkey, make me a {object name}"

The output text gets saved to a scad code, run openscad render to stl.

Show the stl in high resolution in the CLI