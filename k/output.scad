// Cat Model in OpenSCAD
// Parameters
cat_width = 60;
cat_height = 40;
cat_depth = 25;
ear_width = 15;
ear_height = 25;
eye_radius = 2;
nose_radius = 8;
mouth_radius = 10;
tail_length = 50;
// Body
module body() {
hull() {
translate([-cat_width/2, -cat_height/2, -cat_depth/2])
sphere(cat_width/2);
translate([0, cat_height/2, -cat_depth/2])
sphere(cat_height/2);
}
}
// Ears
module ear(angle) {
rotate([0,0,angle])
translate([cat_width/2, cat_height/2, cat_depth/2])
cylinder(r=ear_width, h=ear_height, center=false);
}
// Eyes
module eye(size) {
sphere(size);
}
// Nose
module nose() {
cylinder(r=nose_radius, h=10, center=false);
}
// Mouth
module mouth() {
cylinder(r=mouth_radius, h=2, center=false);
}
// Tail
module tail() {
cylinder(r=cat_width/2, h=tail_length, center=false);
}
// Assembly
difference() {
body();
// Ears
translate([-cat_width/2 + cat_width/4, -cat_height/2 - ear_height/2, -cat_depth/2 - ear_height/2])
ear(30);
translate([-cat_width/2 + cat_width/4, -cat_height/2 - cat_height/4, -cat_depth/2 - ear_height/4])
ear(-30);
// Nose
translate([0, -cat_height/2 - nose_radius, -cat_depth/2 - nose_radius])
nose();
// Eyes
translate([0, cat_height/2 - eye_radius, -cat_depth/2 - eye_radius])
eye(cat_width/4);
translate([0, -cat_height/2 - eye_radius, -cat_depth/2 - eye_radius])
eye(cat_width/4);
// Mouth
translate([0, -cat_height/2 - mouth_radius, -cat_depth/2 - mouth_radius])
mouth();
// Tail
translate([-cat_width/2 - cat_width/4, cat_height/2 - cat_height/4, -cat_depth/2 - cat_depth/4])
tail();
}