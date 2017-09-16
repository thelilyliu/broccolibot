$(document).ready(function() {
    $('html, body').animate({scrollTop:0},'50')
})

$("#banner").on('change', '#imgInp', function(){
    readURL(this)
})

$(".totop").click(function(e){
    $('html, body').animate({scrollTop:0},'50')
})

$("#uploadImages").click(function() {
    //sendfile
    console.log('1')
    var input = document.querySelector('#imgInp')

    uploadImage(input)
})

function parseResponseJSON() {
    $.ajax({
        type: 'GET',
        url: 'http://10.21.38.34:4444/assets/response.json',
        dataType: 'json',
        cache: false
    }).done(function(json, textStatus, jqXHr) {

        var _food = json.images[0].classifiers[0].classes[0].class
        var _food_amount = $("#food-amount").val()

        alert(_food + " " + _food_amount)

    }).fail(function(jqXHr, textStatus, errorThrown) {
        handleAjaxError(jqXHr, textStatus)
    }).always(function() {})
}

function readURL(input) {
    if (input.files && input.files.length) {
        var reader = new FileReader()
        
        reader.onload = function(e) {
            $('#profileBackground').prop('src', e.target.result)
        }
        
        reader.readAsDataURL(input.files[0])
    }
    else {
        $('#profileBackground').prop('src', '')
    }
}

function uploadImage(input) {
    console.log('2')
    var xhr = new XMLHttpRequest()
    var url = 'http://10.21.38.34:4444/postImage'
    var fd = new FormData()
    
    fd.append('uploadFile', input.files[0])       
    xhr.open('POST', url, true)
    
    xhr.onreadystatechange = function() {
        if (xhr.readyState == 4 && xhr.status == 200) { // file upload success
            if (xhr.responseText != 'fail') { // successful upload
                console.log('Success')
            }
            else { // unsuccessful upload
                console.log('You suck bigly')
            }
        }
    }
    
    xhr.send(fd)
}