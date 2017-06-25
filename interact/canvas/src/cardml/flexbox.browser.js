import {AddToClassName, AddToInlineStyle} from '../util/util'
import './flexbox.browser.css'

class FlexboxBrowser {
   
   
   // applyFlexDirection adds the flexbox direction
   applyFlexDirection($el, defaultValue = 'column') {
       const direction = $el.attribs["direction"] || defaultValue
       delete $el.attribs.direction
       $el.attribs["className"] = AddToClassName($el.attribs["className"], `flex-direction-${direction}`)
   }
   
   // applyFlexWrap adds flex wrap 
   applyFlexWrap($el, defaultValue = 'nowrap') {
       const flexWrap = $el.attribs["wrap"] || defaultValue
       delete $el.attribs.wrap
       $el.attribs["className"] = AddToClassName($el.attribs["className"], `flex-wrap-${flexWrap}`)
   }
   
   // applyJustifyContent adds justify-content
   applyJustifyContent($el, defaultValue = 'flex-start') {
       const justifyContent = $el.attribs["justify-content"] || defaultValue
       delete $el.attribs["justify-content"]
       $el.attribs["className"] = AddToClassName($el.attribs["className"], `justify-content-${justifyContent}`)
   }
   
   // applyAlignItems adds align-items
   applyAlignItems($el, defaultValue = 'stretch') {
       const alignItems = $el.attribs["align-items"] || defaultValue
       delete $el.attribs["align-items"]
       $el.attribs["className"] = AddToClassName($el.attribs["className"], `align-items-${alignItems}`)
   }
   
    // applyAlignContent adds align-items
   applyAlignContent($el, defaultValue = 'stretch') {
       const alignContent = $el.attribs["align-content"] || defaultValue
       delete $el.attribs["align-content"]
       $el.attribs["className"] = AddToClassName($el.attribs["className"], `align-content-${alignContent}`)
   }
   
   // applyOrder adds flexbox ordering
   applyOrder($el, defaultValue = 0) {
       const flexOrder = $el.attribs["order"] || defaultValue
       delete $el.attribs["order"]
       $el.attribs["style"] = AddToInlineStyle($el.attribs["style"], `order:${flexOrder}`)
   }
   
   // applyGrow adds flex-grow
   applyGrow($el, defaultValue = 0) {
       const flexGrow = $el.attribs["grow"] || defaultValue
       delete $el.attribs["grow"]
       $el.attribs["style"] = AddToInlineStyle($el.attribs["style"], `flex-grow:${flexGrow}`)
   }
   
   // applyShrink adds flex-shrink
   applyShrink($el, defaultValue = 1) {
       const flexShrink = $el.attribs["shrink"] || defaultValue
       delete $el.attribs["shrink"]
       $el.attribs["style"] = AddToInlineStyle($el.attribs["style"], `flex-shrink:${flexShrink}`)   
   }
   
   // applyBasis adds flex-basis
   applyBasis($el, defaultValue = 'auto') {
       const flexBasis = $el.attribs["basis"] || defaultValue
       delete $el.attribs["basis"]
       $el.attribs["style"] = AddToInlineStyle($el.attribs["style"], `flex-basis:${flexBasis}`)   
   }
   
   // applyAlignItems adds align-self
   applyAlignSelf($el, defaultValue = 'auto') {
       const alignSelf = $el.attribs["align-self"] || defaultValue
       delete $el.attribs["align-self"]
       $el.attribs["className"] = AddToClassName($el.attribs["className"], `align-self-${alignSelf}`)
   }
   
   // flexify takes a cheerio element and parses
   // the attributes to HTML5 flexbox equivalent
   flexify($el) {
       for (const attr in $el.attribs) {
           switch (attr) {
            case 'direction':
                this.applyFlexDirection($el)
                break;
            case 'wrap':
                this.applyFlexWrap($el)
                break;
            case 'justify-content':
                this.applyJustifyContent($el)
                break;
            case 'align-items':
                this.applyAlignItems($el)
                break;
            case 'align-content':
                this.applyAlignContent($el)
                break;
            case 'order':
                this.applyOrder($el)
                break;
            case 'grow':
                this.applyGrow($el)
                break;
            case 'shrink':
                this.applyShrink($el)
                break;
            case 'basis':
                this.applyBasis($el)
                break;
            case 'align-self':
                this.applyAlignSelf($el)
                break;
            case 'style': 
                break;
            default:
                throw new Error(`unknown property '${attr}'`)
           }
       }
   }
}

export default FlexboxBrowser