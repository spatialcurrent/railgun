// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	//"github.com/spatialcurrent/railgun/railgun"
	"net/http"
)

import (
//"github.com/pkg/errors"
)

type HomeHandler struct {
	*BaseHandler
}

func (h *HomeHandler) buildHtmlHead(w http.ResponseWriter, r *http.Request) string {
	return `
  <head>
    <link rel="stylesheet" href="https://cdn.rawgit.com/openlayers/openlayers.github.io/master/en/v5.2.0/css/ol.css" type="text/css">
    <style>
      .map {
        height: 100%;
        width: 100%;
      }
      #info {
        z-index: 1;
        opacity: 0;
        position: absolute;
        bottom: 0;
        left: 0;
        margin: 0;
        background: rgba(0,60,136,0.7);
        color: white;
        border: 0;
        transition: opacity 100ms ease-in;
      }
    </style>
    <script src="https://cdn.rawgit.com/openlayers/openlayers.github.io/master/en/v5.2.0/build/ol.js"></script>
    <title>Railgun</title>
  </head>
  `
}

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	str := "<!doctype html>"
	str += "<html lang=\"en\">"
	str += h.buildHtmlHead(w, r)
	str += `
      <body>
        <div>
          <input
            id="dfl"
            type="text"
            style="width: 80%;font-size:1.4rem;line-height:1.8rem;font-family:Verdana;padding: 10px;">
          <button id="updateLayer" class="block">Button</button>
        </div>
        <div id="map" class="map">
          <pre id="info"/>
        </div>
        <script type="text/javascript">
          var input = document.getElementById("dfl");
          input.value = (new URL(window.location.href)).searchParams.get("dfl")
          input.addEventListener("keyup", function(event) {
            event.preventDefault();
            if (event.keyCode === 13) {
              document.getElementById("updateLayer").click();
            }
          });
          var btn = document.getElementById("updateLayer");
          btn.addEventListener("click", function() {
            let url = new URL(window.location.href);
            
            let maskUrl = createMaskUrl(
              "amenities",
              document.getElementById("dfl").value);
              
            maskLayer.getSource().setUrl(maskUrl);
            maskLayer.getSource().refresh();
            
            let vectorUrl = url.searchParams.get("url") || createVectorUrl(
              "workout_geojson",
              document.getElementById("dfl").value,
              url.searchParams.get("limit"));
            
            vectorLayer.setSource(createVectorSource(vectorUrl));
          });
        </script>
        <script>
          var createMaskUrl = function(layerName, dfl) {
            var maskUrl = "/layers/"+layerName+"/mask/tiles/{z}/{x}/{y}.png";
            var qs = {"threshold": "1", "zoom": "17", "alpha": 200};
            if (dfl != undefined && dfl.length > 0 ) {
              qs.dfl = encodeURI(dfl);
            }
            var keys = Object.keys(qs);
            if(keys.length > 0) {
             maskUrl += "?";
             for(var i = 0; i < keys.length; i++) {
               if(i > 0) {
                 maskUrl += "&"
               }
               maskUrl += keys[i] + "=" + qs[keys[i]];
             }
            };
            return maskUrl;
          };
          
          
          var createVectorUrl = function(serviceName, dfl, limit) {
            var vectorUrl = "/services/"+serviceName+"/tiles/{z}/{x}/{y}.geojson";
            var qs = {"buffer": "1"};
            if (dfl != undefined && dfl.length > 0 ) {
              qs.dfl = encodeURI(dfl);
            }
            if (limit != undefined && limit.length > 0 ) {
              qs.limit = limit;
            }
            var keys = Object.keys(qs);
            if(keys.length > 0) {
             vectorUrl += "?";
             for(var i = 0; i < keys.length; i++) {
               if(i > 0) {
                 vectorUrl += "&"
               }
               vectorUrl += keys[i] + "=" + qs[keys[i]];
             }
            };
            return vectorUrl;
          };
          
          var createVectorSource = function(vectorUrl) {
            return new ol.source.VectorTile({
              format: new ol.format.GeoJSON({"featureProjection": "EPSG:3857", "dataProjection": "EPSG:4326"}),
              maxZoom: 19,
              url: vectorUrl,
              tileLoadFunction: function(tile, url) {
                tile.setLoader(function() {
                  var xhr = new XMLHttpRequest();
                  xhr.open('GET', url);
                  xhr.onload = function() {
                    var json = JSON.parse(xhr.responseText);
                    json.features.forEach((f) => {
                      f.properties._tile_z = tile.tileCoord[0];
                      f.properties._tile_x = tile.tileCoord[1];
                      f.properties._tile_y = tile.tileCoord[2];
                    });
                    var format = tile.getFormat();
                    tile.setFeatures(format.readFeatures(json));
                    tile.setProjection(format.defaultFeatureProjection);
                  }
                  xhr.send();
                });
              }
            });
          };
        </script>
        <script type="text/javascript">

          var baseLayer = new ol.layer.Tile({
            source: new ol.source.OSM()
          });
          
          let url = new URL(window.location.href);
          
          var maskUrl = createMaskUrl(
            "amenities",
            url.searchParams.get("dfl"));
          
          var vectorUrl = url.searchParams.get("url") || createVectorUrl(
            "workout_geojson",
            url.searchParams.get("dfl"),
            url.searchParams.get("limit"));
            
          var vectorSource = createVectorSource(vectorUrl);
          
          var maskLayer = new ol.layer.Tile({
            source: new ol.source.XYZ({
              url: maskUrl
            })
          });
          
          var vectorLayer = new ol.layer.VectorTile({
            source: vectorSource,
            renderBuffer: 256,
            renderMode: "image",
            style: function(feature) {
              return new ol.style.Style({
                image: new ol.style.Circle({
                  radius: 8,
                  fill: new ol.style.Fill({color:'rgba(0,0,255,0.8)'}),
                  stroke: new ol.style.Stroke({color: 'black', width: 2})
                })
              });
            }
          });
          
          //var layers = [baseLayer, maskLayer, vectorLayer];
          var layers = [baseLayer, vectorLayer];
          //var layers = [baseLayer, maskLayer];
          
          var map = new ol.Map({
            target: 'map',
            layers: layers,
            view: new ol.View({
              center: ol.proj.fromLonLat([0, 0]),
              zoom: 3
            })
          });
              
          map.on('pointermove', function(event) {
            let info = document.getElementById('info')
            var features = map.getFeaturesAtPixel(event.pixel);
            if (!features) {
              info.innerText = '';
              info.style.opacity = 0;
              return;
            }
            var f = features[0];
            //var properties = features[0].getProperties();
            var properties = f.getKeys()
              .filter(x => x != "geometry")
              .map(x => [x, f.get(x)])
              .reduce(function(o, x) { o[x[0]] = x[1]; return o; }, {});
            info.innerText = JSON.stringify(properties, null, 2);
            info.style.opacity = 1;
          });
          
        </script>
      </body>
    </html>
  `

	w.Write([]byte(str)) // #nosec

}
