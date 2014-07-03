'use strict';

//http://stackoverflow.com/questions/979975/how-to-get-the-value-from-url-parameter
var QueryString = function () {
  // This function is anonymous, is executed immediately and 
  // the return value is assigned to QueryString!
  var query_string = {};
  var query = window.location.search.substring(1);
  var vars = query.split("&");
  for (var i=0;i<vars.length;i++) {
    var pair = vars[i].split("=");
      // If first entry with this name
    if (typeof query_string[pair[0]] === "undefined") {
      query_string[pair[0]] = pair[1];
      // If second entry with this name
    } else if (typeof query_string[pair[0]] === "string") {
      var arr = [ query_string[pair[0]], pair[1] ];
      query_string[pair[0]] = arr;
      // If third or later entry with this name
    } else {
      query_string[pair[0]].push(pair[1]);
    }
  } 
    return query_string;
} ();



// Declare app level module which depends on filters, and services
var RtRt = angular.module('RtRt', ['$strap.directives', 'ngResource','RtRt.filters'])

RtRt.config(['$httpProvider', function($httpProvider) {

      // Use x-www-form-urlencoded Content-Type
      $httpProvider.defaults.headers.post['Content-Type'] = 'application/x-www-form-urlencoded;charset=utf-8';
          
      // Override $http service's default transformRequest
      $httpProvider.defaults.transformRequest = [function(data)
      {
        return angular.isObject(data) && String(data) !== '[object File]' ? jQuery.param(data) : data;
      }];


  }]);


RtRt.run(['$location', '$rootScope', '$http', function($location, $rootScope, $http) {
    $rootScope.twitter = JSON.parse(twitterObj);
    $rootScope.retweets = JSON.parse(retweetObj);
    $rootScope.start = 0;
    $rootScope.updateStart = function(n){
      console.log(n);
      $rootScope.start = 0;
      console.log($rootScope.retweets.length)
      if(n >= 0 && n <= $rootScope.retweets.length) {
        console.log(1)
        if(n == 5){
        console.log(2)
          n = 4;
        } 
        console.log(3)
        $rootScope.start = n;
      }
    }
    console.log($rootScope.twitter);
    console.log($rootScope.retweets);
    $rootScope.updateRadios = function(){
      console.log('updating radios...')
      setupLabel();
      $rootScope.share_btn = "Share";

    };
    $rootScope.selectedTweet = '';
    $rootScope.share_btn = "Share";
    $rootScope.retractionPlaceholder = 'ex. #RT<<'
    $rootScope.shareText = "RT<< '@example : fobar' http://rtrt.co/r/1234 via @retwact "
    $rootScope.updateSelectedTweet = function(t){
        $rootScope.share_btn = "Share";
        document.getElementById('retractionBox').disabled=false;
        $rootScope.selectedTweet = t.id_str;
        $rootScope.selectedTweetTXT = t.text;
        $rootScope.retractionPlaceholder = 'RT<< '+$rootScope.selectedTweetTXT
        $rootScope.retractionText = 'RT<<  '+$rootScope.selectedTweetTXT
        $rootScope.shareText = "RT<< '@"+$rootScope.twitter.screen_name+" : "+$rootScope.selectedTweetTXT.slice(0,50)+"...' http://rtrt.co/r/"+$rootScope.selectedTweet+" via @retwact"
        $rootScope.shareText2 = "RT<< '"+$rootScope.selectedTweetTXT.slice(0,50)+"...' http://rtrt.co/r/"+$rootScope.selectedTweet+" via @retwact"
        $rootScope.shareLink = 'http://twitter.com/home/?status='+$rootScope.shareText;
        $rootScope.shareLink2 = 'http://twitter.com/home/?status='+$rootScope.shareText2;
        $(".toggle").each(function(index, toggle) {
            toggleHandler(toggle);
        });

    }
    $rootScope.shareLink = "#";

    $rootScope.updateRetractionShare = function(){
      $rootScope.retractionText = $rootScope.retractionText.slice(0,50)
    
      $rootScope.shareText = ""+$rootScope.retractionText.slice(0,50)+" http://rtrt.co/r/"+$rootScope.selectedTweet+" via @retwact"
      $rootScope.shareText2 = ""+$rootScope.retractionText.slice(0,50)+" http://rtrt.co/r/"+$rootScope.selectedTweet+" via @retwact"
      $rootScope.shareLink = 'http://twitter.com/home/?status='+$rootScope.shareText;
      $rootScope.shareLink2 = 'http://twitter.com/home/?status='+$rootScope.shareText2;
    }
    $rootScope.deleteTweet = false;
    $rootScope.setDeleteTweet = function(d){
      if(d == 2){
        $rootScope.deleteTweet = true;
      } else {
        $rootScope.deleteTweet = false;        
      }
      console.log($rootScope.deleteTweet);

    }
    $rootScope.share = function(){
      console.log('sharing is caring...')
      if($rootScope.share_btn == "Share"){

      var p = {}
      p.tid = $rootScope.selectedTweet;
      p.uid = $rootScope.retweets[0].user.screen_name;
      p.rm = $rootScope.shareText;
      p.orm = $rootScope.selectedTweetTXT;
      p.verf = QueryString.oauth_verifier;
      p.del = $rootScope.deleteTweet;
      console.log(p)
      if(p.orm){
      $http.post('/link',p).success(function(data){
        console.log(data);
        if(data == 'success'){
          $rootScope.share_btn = "Retwaction Successfully Launched!"
          $rootScope.selectedTweet = false;
          window.open($rootScope.shareLink,'_blank', 'toolbar="no", height=275, location="no"')
          //window.location.href = $rootScope.shareLink2;
          return true;

        }
      });

      } else {return false;}
    } else {
      return false
    }
    }
    //$rootScope.share()
/*
    if($rootScope.twitter.user){
      $rootScope.gravatar = "<img class='avatar' src='"+$rootScope.twitter.user.profile_image_url+"'>"
    }
*/

  }]);

angular.module('RtRt.filters', []).
  filter('range', function() {
    return function(input, total) {
      total = parseInt(total);
      for (var i=0; i<=total; i++)
        input.push(i);
      return input;
    };
  }).filter('startFrom', function() {
    //http://jsfiddle.net/2ZzZB/56/
    return function(input, start) {
        start = +start; //parse to int
        return input.slice(start);
    }
});
function Global($scope, $http){ 
  $scope.test2 = "Hello World!"
}; 