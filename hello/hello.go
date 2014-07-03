package hello

import (
    "appengine"
    "appengine/datastore"
    "appengine/user"
    "appengine/memcache"
 //   "appengine/taskqueue"

//    "encoding/json"
//    "errors"
    //"fmt"
    "io/ioutil"
    "html/template"
    "net/http"
    "time"
    "oauth"
    "encoding/json"
)

type Greeting struct {
    Author  string
    Content string
    Date    time.Time
}

type StoredTweet struct {
    Tid string
    Uid string
    Rmessage string
    OrigMessage string
    Date    time.Time
}

type TwitterJSON struct {
    User        string
    Retweets    string
}

func init() {
    http.HandleFunc("/", root)
    http.HandleFunc("/r/", r)
    http.HandleFunc("/link", link)
    http.HandleFunc("/sign", sign)
    http.HandleFunc("/connect", connect)
    http.HandleFunc("/twerk", twerk)

}

const (
        // Created at http://code.google.com/apis/console, these identify
        // our app for the OAuth protocol.
        consumerKey = ""
        consumerSecret = ""
)

var rtoken *oauth.RequestToken
var o *oauth.Consumer
func connect(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    o := oauth.NewConsumer(
        c,
        consumerKey,
        consumerSecret,
        oauth.ServiceProvider{
            RequestTokenUrl:   "https://api.twitter.com/oauth/request_token",
            AuthorizeTokenUrl: "https://api.twitter.com/oauth/authorize",
            AccessTokenUrl:    "https://api.twitter.com/oauth/access_token",
        })
    rtoken, url, err := o.GetRequestTokenAndUrl("http://rtrt.co")
    c.Infof(" url: %v", url)
    c.Infof(" token: %v", rtoken)
    c.Infof(" err: %v", err)


    // Create an Item
    item := &memcache.Item{
        Key:   rtoken.Token,
        Value: []byte(rtoken.Secret),
    }
    // Add the item to the memcache, if the key does not already exist
    if err := memcache.Add(c, item); err == memcache.ErrNotStored {
        c.Infof("item with key %q already exists", item.Key)
    } else if err != nil {
        c.Errorf("error adding item: %v", err)
    }

    http.Redirect(w, r, url, http.StatusFound)
}


func root(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    requestToken := r.FormValue("oauth_token") 
    verificationCode := r.FormValue("oauth_verifier") 
            c.Infof(" code: %v", verificationCode)
            c.Infof(" token_key: %v", requestToken)
            c.Infof(" request: %v", r)
    // Get the item from the memcache



    if item, err := memcache.Get(c, requestToken); err == memcache.ErrCacheMiss {
        c.Infof("item not in the cache")
        renderTemplate(w, "index", &TwitterJSON{User: "[]", Retweets: "[]"})
    } else if err != nil {
        c.Errorf("error getting item: %v", err)
        renderTemplate(w, "index", &TwitterJSON{User: "[]", Retweets: "[]"})
    } else {
        c.Infof("the secret is %q", item.Value)
        secret := string(item.Value)
        rtoken := oauth.RequestToken{Token: requestToken, Secret: secret}
        c.Infof(" rtoken: %v", rtoken)
    
    o := oauth.NewConsumer(
        c,
        consumerKey,
        consumerSecret,
        oauth.ServiceProvider{
            RequestTokenUrl:   "https://api.twitter.com/oauth/request_token",
            AuthorizeTokenUrl: "https://api.twitter.com/oauth/authorize",
            AccessTokenUrl:    "https://api.twitter.com/oauth/access_token",
        })
    
        accessToken, err := o.AuthorizeToken(rtoken, verificationCode)
        if err != nil {
            c.Infof(" err: %v", err)
            http.Redirect(w, r, "/", http.StatusFound)

        } else {
            c.Infof(" accessToken: %v", accessToken)
             // Create an Item
             item1 := &memcache.Item{
                 Key:   verificationCode+"-token",
                 Value: []byte(accessToken.Token),
             }
             // Add the item to the memcache, if the key does not already exist
             if err := memcache.Add(c, item1); err == memcache.ErrNotStored {
                 c.Infof("item with key %q already exists", item1.Key)
             } else if err != nil {
                 c.Errorf("error adding item: %v", err)
             }
              // Create an Item
             item2 := &memcache.Item{
                 Key:   verificationCode+"-secret",
                 Value: []byte(accessToken.Secret),
             }
             // Add the item to the memcache, if the key does not already exist
             if err := memcache.Add(c, item2); err == memcache.ErrNotStored {
                 c.Infof("item with key %q already exists", item2.Key)
             } else if err != nil {
                 c.Errorf("error adding item: %v", err)
             }
            response, err := o.Get(
//                "http://api.twitter.com/1/statuses/retweets_of_me.json",
                "https://api.twitter.com/1.1/account/settings.json",
                map[string]string{"count": "1"},
                accessToken)
            if err != nil {
            c.Errorf("error getting tweets: %v", err)

            }
            defer response.Body.Close()
        
            bits, err := ioutil.ReadAll(response.Body)

            c.Infof(" user: %v", string(bits))
            userJSON := string(bits)

            response2, err := o.Get(
                "https://api.twitter.com/1.1/statuses/retweets_of_me.json",
//                "https://api.twitter.com/1.1/statuses/user_timeline.json",
                map[string]string{"count": "10000"},
                accessToken)
            if err != nil {
            c.Errorf("error getting tweets: %v", err)

            }
            defer response2.Body.Close()
        
            bits2, err := ioutil.ReadAll(response2.Body)

            c.Infof(" recent retweets: %v", string(bits2))
            retweetJSON := string(bits2)
            tJSON := &TwitterJSON{ User: userJSON, Retweets: retweetJSON }

            renderTemplate(w, "rtrt", tJSON)
        }
    } 

}

var templates = template.Must(template.ParseFiles("index.html","rtrt.html","link.html","r.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *TwitterJSON) {
      err := templates.ExecuteTemplate(w, tmpl+".html", p)
      if err != nil {
          http.Error(w, err.Error(), http.StatusInternalServerError)
     }
  }
func renderTemplate2(w http.ResponseWriter, tmpl string, p *StoredTweet) {
      err := templates.ExecuteTemplate(w, tmpl+".html", p)
      if err != nil {
          http.Error(w, err.Error(), http.StatusInternalServerError)
     }
  }


var tweetTemplate = template.Must(template.New("tweets").Parse(tweetTemplateHTML))
var stringTemplate = template.Must(template.New("string").Parse(stringTemplateHTML))

const stringTemplateHTML = `{{.}}`
const tweetTemplateHTML = `
<html>
  <head>
    <script type="text/javascript">
        tweets = JSON.parse({{.}})
    </script>
  </head>
</html>
`

var guestbookTemplate = template.Must(template.New("book").Parse(guestbookTemplateHTML))

const guestbookTemplateHTML = `
<html>
  <body>
  "THIS IS A BETA SITE TO MEASURE INTEREST WHILE WE BUILD RTRETRACT.THE FORM WILL NOT RESULT IN A RETWETT RETRACTION UNTIL WE LAUNCH THE LIVE SITE."
 <a href="/connect"><image src="https://dev.twitter.com/sites/default/files/images_documentation/sign-in-with-twitter-link.png"></a>
<br> 
    <iframe src="https://docs.google.com/spreadsheet/embeddedform?formkey=dG1VaHBra3JUcWN0RTBOVElvRFZud2c6MQ" width="100%" height="100%" frameborder="0" marginheight="0" marginwidth="0">Loading...</iframe>
  </body>
</html>
`

const templateHTML = `
<html>
<body class="ss-base-body" dir="ltr" itemscope="" itemtype="http://schema.org/CreativeWork/FormObject" marginwidth="0" marginheight="0"><meta itemprop="name" content="RT Retract">
<meta itemprop="description" content="Retweet Retract provides a service for notifying twitter users of corrections to tweets that they have retweeted and allowing them to retweet the correction.">

<meta itemprop="url" content="https://docs.google.com/spreadsheet/viewform?formkey=dG1VaHBra3JUcWN0RTBOVElvRFZud2c6MQ">
<meta itemprop="embedUrl" content="https://docs.google.com/spreadsheet/embeddedform?formkey=dG1VaHBra3JUcWN0RTBOVElvRFZud2c6MQ">
<meta itemprop="faviconUrl" content="//ssl.gstatic.com/docs/spreadsheets/forms/favicon_jfk2.png">

<div class="ss-form-container"><div class="ss-form-heading"><h1 class="ss-form-title">RT Retract</h1>
<p></p>
<div class="ss-form-desc ss-no-ignore-whitespace">Retweet Retract provides a service for notifying twitter users of corrections to tweets that they have retweeted and allowing them to retweet the correction.</div>

<hr class="ss-email-break" style="display:none;">
<div class="ss-required-asterisk">* Required</div></div>
<div class="ss-form"><form action="https://docs.google.com/spreadsheet/formResponse?formkey=dG1VaHBra3JUcWN0RTBOVElvRFZud2c6MQ&amp;embedded=true&amp;ifq" method="POST" id="ss-form">


<br>
<div class="errorbox-good">
<div class="ss-item ss-item-required ss-text"><div class="ss-form-entry"><label class="ss-q-title" for="entry_0">Twitter Handle
<span class="ss-required-asterisk">*</span></label>
<a href="/connect"><image src="https://dev.twitter.com/sites/default/files/images_documentation/sign-in-with-twitter-link.png"></a>
<br> <div class="errorbox-good">
<div class="ss-item ss-item-required ss-paragraph-text"><div class="ss-form-entry"><label class="ss-q-title" for="entry_1">Original Tweet
<span class="ss-required-asterisk">*</span></label>
<label class="ss-q-help" for="entry_1">The tweet you want to retract</label>
<textarea name="entry.1.single" rows="8" cols="75" class="ss-q-long" id="entry_1"></textarea></div></div></div>
<br> <div class="errorbox-good">
<div class="ss-item  ss-paragraph-text"><div class="ss-form-entry"><label class="ss-q-title" for="entry_2">Retraction or corrected tweet
</label>
<label class="ss-q-help" for="entry_2"></label>
<textarea name="entry.2.single" rows="8" cols="75" class="ss-q-long" id="entry_2"></textarea></div></div></div>
<br>
<input type="hidden" name="pageNumber" value="0">
<input type="hidden" name="backupCache" value="">


<div class="ss-item ss-navigate"><div class="ss-form-entry">
<input type="submit" name="submit" value="Submit">
<div class="password-warning">Never submit passwords through Google Forms.</div></div></div></form>
</div>
<div class="ss-footer"><div class="ss-attribution"></div>
<div class="ss-legal"><span class="ss-powered-by">Powered by <a href="http://docs.google.com">Google Docs</a></span>
<span class="ss-terms"><small><a href="https://docs.google.com/spreadsheet/reportabuse?formkey=dG1VaHBra3JUcWN0RTBOVElvRFZud2c6MQ&amp;source=https://docs.google.com/spreadsheet/embeddedform?formkey%3DdG1VaHBra3JUcWN0RTBOVElvRFZud2c6MQ">Report Abuse</a>
-
<a href="http://www.google.com/accounts/TOS">Terms of Service</a>
-
<a href="http://www.google.com/google-d-s/terms.html">Additional Terms</a></small></span></div></div></div></body>
</html>
`
func link(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    o := oauth.NewConsumer(
        c,
        consumerKey,
        consumerSecret,
        oauth.ServiceProvider{
            RequestTokenUrl:   "https://api.twitter.com/oauth/request_token",
            AuthorizeTokenUrl: "https://api.twitter.com/oauth/authorize",
            AccessTokenUrl:    "https://api.twitter.com/oauth/access_token",
    })
    if r.FormValue("tid") != "" {
        tid := r.FormValue("tid")
        trm := StoredTweet{
            Tid: r.FormValue("tid"),
            Uid: r.FormValue("uid"),
            Rmessage: r.FormValue("rm"),
            OrigMessage: r.FormValue("orm"),
            Date:    time.Now(),
        }
                c.Infof("delete tweet %q", r.FormValue("del"))

        rmsg := r.FormValue("rm")
        verificationToken := r.FormValue("verf")+"-token"
        verificationSecret := r.FormValue("verf")+"-secret"
        _, err := datastore.Put(c, datastore.NewIncompleteKey(c, "Retracts", nil), &trm)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        //my_t := taskqueue.NewPOSTTask("/twerk", map[string][]string{"twitter_user": {""}, "retract_message":{rmsg},})
        //if _, err := taskqueue.Add(c, my_t, ""); err != nil {
        //    http.Error(w, err.Error(), http.StatusInternalServerError)
        //    return
        //}

        if verificationToken != "-token" {
            if token, err := memcache.Get(c, verificationToken); err == memcache.ErrCacheMiss {
                c.Infof("token not in the cache")
            } else if err != nil {
                c.Errorf("error getting item: %v", err)
            } else {
                c.Infof("the token is %q", token.Value)
                if secret, err := memcache.Get(c, verificationSecret); err == memcache.ErrCacheMiss {
                    c.Infof("secret not in the cache")
                } else if err != nil {
                    c.Errorf("error getting item: %v", err)
                } else {
                    c.Infof("the secret is %q", secret.Value)
    
                   // token := "1374797204-aCb29F5o3i9a7DCKAQ9BooaIrMccVrstZ2P47dv"
                   // secret := "hrZfqa0jGcboHVHS8m7XnF0H9qibgUx6oEVSwyklM"
                    accessToken := &oauth.AccessToken{ Token : string(token.Value), Secret : string(secret.Value)}
                            c.Infof(" tweetid : %v", tid)
            
                    response, err := o.Get("https://api.twitter.com/1/statuses/retweets/"+string(tid)+".json",
                            map[string]string{"count": "3"},
                            accessToken)
                        c.Infof("url used: https://api.twitter.com/1.1/statuses/retweets/%v.json", tid)
                    if err != nil {
                        c.Errorf("error getting tweets: %v", err)
                    }
                    defer response.Body.Close()
                    
                    bits, err := ioutil.ReadAll(response.Body)
                    //http://stackoverflow.com/questions/6581575/golang-help-reflection-to-get-values
                    //c.Infof(" retweets : %v", string(bits))
                    jsonStr := string(bits)
                    var f interface{}
                    b := []byte(jsonStr)
                    err2 := json.Unmarshal(b, &f)
                    if err2 != nil {
                        c.Errorf("unable to decode json")
                    } 
                    m := f.([]interface{})
                    for _,v := range m {
                        //rt := reflect.ValueOf(v)
                        //user := rt.Field(0)
                        rt := v.(map[string]interface{})
                        //id := strconv.Itoa(rt["id"].(int))
                        user := rt["user"].(map[string]interface{})
                        screenname := user["screen_name"]
                        //c.Infof(" retweetid: %v", id)
                        c.Infof(" retweeter: %v", screenname)
                        c.Infof(" retweet: %v", rmsg)
            
                
                        /*if screenname != "" {
                            t := taskqueue.NewPOSTTask("/twerk", map[string][]string{"twitter_user": {screenname.(string)}, "retract_message":{rmsg},})
                            if _, err := taskqueue.Add(c, t, ""); err != nil {
                                http.Error(w, err.Error(), http.StatusInternalServerError)
                                return
                            }
                        }*/
                    }
                    if r.FormValue("del") == "true" {
                        response1, err := o.Post("https://api.twitter.com/1/statuses/destroy/"+string(tid)+".json",
                            map[string]string{},
                            accessToken)
                        if err != nil {
                            c.Errorf("error deleting tweet: %v", err)
                        }
                        defer response1.Body.Close()
                    
                        bits1, err := ioutil.ReadAll(response1.Body)
                        c.Infof(" delete tweet result: %v", string(bits1))

                    }

                }

            }
        }
        


        if err :=  stringTemplate.Execute(w, "success"); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
    } else {
        if err :=  stringTemplate.Execute(w, "failed"); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
    }
}

func r(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    tid := r.URL.Path[3:]
    if tid != "" {
        c.Infof(" tid: %v", tid)

        q := datastore.NewQuery("Retracts").Filter("Tid =", tid).Limit(1)
        retract := make([]StoredTweet, 0, 1)
        if _, err := q.GetAll(c, &retract); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        if len(retract) < 1 {
            http.Redirect(w, r, "/", http.StatusFound)
        } else {
            ret := retract[0]
            renderTemplate2(w, "r", &ret);
        }
    } else {
        http.Redirect(w, r, "/", http.StatusFound)
    }
}

func sign(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    if r.FormValue("content") != "1" {
    	g := Greeting{
        	Content: r.FormValue("content"),
        	Date:    time.Now(),
    	}

    	if u := user.Current(c); u != nil {
        	g.Author = u.String()
    	}
    	_, err := datastore.Put(c, datastore.NewIncompleteKey(c, "Greeting", nil), &g)
    	if err != nil {
        	http.Error(w, err.Error(), http.StatusInternalServerError)
        	return
    	}
    }
    http.Redirect(w, r, "/", http.StatusFound)
}

func twerk(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    o := oauth.NewConsumer(
        c,
        consumerKey,
        consumerSecret,
        oauth.ServiceProvider{
            RequestTokenUrl:   "https://api.twitter.com/oauth/request_token",
            AuthorizeTokenUrl: "https://api.twitter.com/oauth/authorize",
            AccessTokenUrl:    "https://api.twitter.com/oauth/access_token",
    })
    tuser := r.FormValue("twitter_user")
    rtmsg := r.FormValue("retract_message")
    //rplyid := r.FormValue("in_reply_to_status_id")
    //token := "1394797376-C5ayPtPMpAqVoXb76QM2zob52fZOHzgUjkoebCC"
    //secret := "BKJFMIhHNCRC58L2ePAKH2LZKXM4B12s9elTInRutk"
    token := "1374797204-aCb29F5o3i9a7DCKAQ9BooaIrMccVrstZ2P47dv"
    secret := "hrZfqa0jGcboHVHS8m7XnF0H9qibgUx6oEVSwyklM"
    accessToken := &oauth.AccessToken{ Token : token, Secret : secret}
        c.Infof(" tuser : %v", tuser)
        c.Infof(" rtmsg : %v", rtmsg)

    if tuser != "" && rtmsg != "" {
        msg := "@"+tuser+" "+rtmsg
        response, err := o.Post("https://api.twitter.com/1.1/statuses/update.json",
//                map[string]string{"status": msg, "in_reply_to_status_id": rplyid,},
                map[string]string{"status": msg, },
                accessToken)
        if err != nil {
            c.Errorf("error getting tweets: %v", err)
        }
        defer response.Body.Close()
        
        bits, err := ioutil.ReadAll(response.Body)

        c.Infof(" tweet bcast resp : %v", string(bits))
    }
    if tuser == "" && rtmsg != "" {
        msg := rtmsg
        response, err := o.Post("https://api.twitter.com/1.1/statuses/update.json",
//                map[string]string{"status": msg, "in_reply_to_status_id": rplyid,},
                map[string]string{"status": msg, },
                accessToken)
        if err != nil {
            c.Errorf("error getting tweets: %v", err)
        }
        defer response.Body.Close()
        
        bits, err := ioutil.ReadAll(response.Body)

        c.Infof(" tweet bcast resp : %v", string(bits))
    }
    /*key := datastore.NewKey(c, "Counter", name, 0, nil)
    var counter Counter
    if err := datastore.Get(c, key, &counter); err == datastore.ErrNoSuchEntity {
        counter.Name = name
    } else if err != nil {
        c.Errorf("%v", err)
        return
    }
    counter.Count++
    if _, err := datastore.Put(c, key, &counter); err != nil {
        c.Errorf("%v", err)
    }*/
}

