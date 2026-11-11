var studentId, x;
        //here is the object carrying the students data
    studentId = [{
        // password: document.getElementById('id01').style.display='none'
        password: 'typewriterChange.html'
    }]

var modal = document.getElementById('id01');
// When the user clicks anywhere outside of the modal, close it
window.onclick = function(event) {
    if (event.target == modal) {
        modal.style.display = "none";
    }
}

var type = [
    'I am <b>Niksar</b>',
    'I like to build things',
    'I have a few discord bots',
    'I also like to create things in photoshop',
    "I ^1000\n <strong>am Niksar</strong>"
]


var twChange = (x) => {
    var text1 = document.forms["theForm3"]["type1"].value;
    var text2 = document.forms["theForm3"]["type2"].value;
    var text3 = document.forms["theForm3"]["type3"].value;
    var text4 = document.forms["theForm3"]["type4"].value;
    var text5 = document.forms["theForm3"]["type5"].value;
    document.getElementById('no1').innerHTML = text1;
    type[0] = text1;
    type[1] = text2;
    type[2] = text3;
    type[3] = text4;
    type[4] = text5;
    console.log(type[0]);
    console.log(text2);
    console.log(text3);
    console.log(text4);
    console.log(text5);
}