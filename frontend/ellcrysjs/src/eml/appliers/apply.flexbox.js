// @flow
import {AddToClassName, AddToInlineStyle} from '../../util/util'
import cheerio from 'cheerio'

export type FlexboxApplier = {
    applyFlexDirection: ($el: cheerio.CheerioElement, defaultValue?: string) => void,
    applyFlexWrap: ($el: cheerio.CheerioElement, defaultValue?: string) => void,
    applyJustifyContent: ($el: cheerio.CheerioElement, defaultValue?: string) => void,
    applyAlignItems: ($el: cheerio.CheerioElement, defaultValue?: string) => void,
    applyAlignContent: ($el: cheerio.CheerioElement, defaultValue?: string) => void,
    applyOrder: ($el: cheerio.CheerioElement, defaultValue?: number) => void,
    applyGrow: ($el: cheerio.CheerioElement, defaultValue?: number) => void,
    applyShrink: ($el: cheerio.CheerioElement, defaultValue?: number) => void,
    applyBasis: ($el: cheerio.CheerioElement, defaultValue?: string) => void,
    applyAlignSelf: ($el: cheerio.CheerioElement, defaultValue?: string) => void,
}

export const Appliers: FlexboxApplier = {
    
    // applyFlexDirection adds the flexbox direction
   applyFlexDirection: ($el: cheerio.CheerioElement, defaultValue: string = 'column'): void => {
       const direction = ($el.attribs["direction"]: any) || defaultValue
       delete $el.attribs.direction
       $el.attribs["className"] = AddToClassName($el.attribs["className"], `flex-direction-${direction}`)
   },
   
   // applyFlexWrap adds flex wrap 
   applyFlexWrap: ($el: cheerio.CheerioElement, defaultValue: string = 'nowrap'): void => {
       const flexWrap = $el.attribs["wrap"] || defaultValue
       delete $el.attribs.wrap
       $el.attribs["className"] = AddToClassName($el.attribs["className"], `flex-wrap-${flexWrap}`)
   },
   
   // applyJustifyContent adds justify-content
   applyJustifyContent: ($el: cheerio.CheerioElement, defaultValue: string = 'flex-start'): void => {
       const justifyContent = $el.attribs["justifycontent"] || defaultValue
       delete $el.attribs["justifycontent"]
       $el.attribs["className"] = AddToClassName($el.attribs["className"], `justify-content-${justifyContent}`)
   },
   
   // applyAlignItems adds align-items
   applyAlignItems: ($el: cheerio.CheerioElement, defaultValue: string = 'stretch'): void => {
       const alignItems = $el.attribs["alignitems"] || defaultValue
       
       delete $el.attribs["alignitems"]
       $el.attribs["className"] = AddToClassName($el.attribs["className"], `align-items-${alignItems}`)
   },
   
    // applyAlignContent adds align-items
   applyAlignContent: ($el: cheerio.CheerioElement, defaultValue: string = 'stretch'): void => {
       const alignContent = $el.attribs["aligncontent"] || defaultValue
       delete $el.attribs["aligncontent"]
       $el.attribs["className"] = AddToClassName($el.attribs["className"], `align-content-${alignContent}`)
   },
   
   // applyOrder adds flexbox ordering
   applyOrder: ($el: cheerio.CheerioElement, defaultValue: number = 0): void => {
       const flexOrder = $el.attribs["order"] || defaultValue
       delete $el.attribs["order"]
       $el.attribs["style"] = AddToInlineStyle($el.attribs["style"], `order:${flexOrder}`)
   },
   
   // applyGrow adds flex-grow
   applyGrow: ($el: cheerio.CheerioElement, defaultValue: number = 0): void => {
       const flexGrow = $el.attribs["grow"] || defaultValue
       delete $el.attribs["grow"]
       $el.attribs["style"] = AddToInlineStyle($el.attribs["style"], `flex-grow:${flexGrow}`)
   },
   
   // applyShrink adds flex-shrink
   applyShrink: ($el: cheerio.CheerioElement, defaultValue: number = 1): void => {
       const flexShrink = $el.attribs["shrink"] || defaultValue
       delete $el.attribs["shrink"]
       $el.attribs["style"] = AddToInlineStyle($el.attribs["style"], `flex-shrink:${flexShrink}`)   
   },
   
   // applyBasis adds flex-basis
   applyBasis: ($el: cheerio.CheerioElement, defaultValue: string = 'auto'): void => {
       const flexBasis = $el.attribs["basis"] || defaultValue
       delete $el.attribs["basis"]
       $el.attribs["style"] = AddToInlineStyle($el.attribs["style"], `flex-basis:${flexBasis}`)   
   },
   
   // applyAlignItems adds align-self
   applyAlignSelf: ($el: cheerio.CheerioElement, defaultValue: string = 'auto'): void => {
       const alignSelf = $el.attribs["alignself"] || defaultValue
       delete $el.attribs["alignself"]
       $el.attribs["className"] = AddToClassName($el.attribs["className"], `align-self-${alignSelf}`)
   }
}