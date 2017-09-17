$(document).ready(function() {
    $('html, body').animate({scrollTop:0},'50')
})

$("#banner").on('change', '#imgInp', function(){
    readURL(this)
})

$(".totop").click(function(e){
    $('html, body').animate({scrollTop:0},'50')
})

var servingsize=0

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
    var url = '/postImage'
    var fd = new FormData()
    
    fd.append('uploadFile', input.files[0])       
    xhr.open('POST', url, true)
    
    xhr.onreadystatechange = function() {
        if (xhr.readyState == 4 && xhr.status == 200) { // file upload success
            if (xhr.responseText != 'fail') { // successful upload
                console.log('Success')
                servingsize = $("#food-amount").val()
                alert(servingsize)
            }
            else { // unsuccessful upload
                console.log('You suck bigly')
            }
        }
    }
    
    xhr.send(fd)
}

$("#uploadServing").click(function(e){
    e.preventDefault()
    $.ajax({
        type: 'GET',
        url: '/assets/response.json',
        dataType: 'json',
        cache: false
    }).done(function(json, textStatus, jqXHr) {

        var _food = json.images[0].classifiers[0].classes[0].class
        var _food_amount = $("#food-amount").val()

        $("#form-foodserving").hide()
        $("#food-display").removeClass("hidden")

        calcNutrition(_food, _food_amount)

    }).fail(function(jqXHr, textStatus, errorThrown) {
    }).always(function() {})
})

function calcNutrition(foodname, foodamount) {
    console.log(foodname, foodamount)
    $.ajax({
        type: 'POST',
        url: '/calculateNutrition/' + foodname.toLowerCase() + '/' + foodamount.toLowerCase(),
        dataType: 'json',
        cache: false
    }).done(function(json, textStatus, jqXHr) {
        console.log(json)
        $("#table-calories").val(json.Calories)
        $("#table-fat").val(json.Fat)
        $("#table-soudium").val(json.Sodium)
        $("#table-cholesterol").val(json.Cholesterol)
        $("#table-protein").val(json.Protein)
        $("#table-carbohydrates").val(json.Carbohydrates)

        $("#table-calories").val(json.calories)
        $("#table-fat").val(json.fat)
        $("#table-soudium").val(json.sodium)
        $("#table-cholesterol").val(jsoncCholesterol)
        $("#table-protein").val(json.protein)
        $("#table-carbohydrates").val(json.carbohydrates)

    }).fail(function(jqXHr, textStatus, errorThrown) {
    }).always(function() {})
}