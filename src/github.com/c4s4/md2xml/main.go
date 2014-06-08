package main

import (
    "os"
    "github.com/russross/blackfriday"
    "os/exec"
    "io/ioutil"
    "fmt"
    "strings"
)

const (
    HELP = `md2xml [-h] file.md
Transform a given Markdown file into XML.
-h        To print this help page.
file.md   The markdown file to convert.
Note: this program calls xsltproc that must have been installed.`
    STYLESHEET = `<?xml version="1.0" encoding="utf-8"?>
<!--
Stylesheet to transform an XHTML document to XML one.
-->

<xsl:stylesheet xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
                version="1.0">

  <xsl:output method="xml" encoding="UTF-8"/>
  <xsl:param name="id">ID</xsl:param>
  <xsl:param name="date">DATE</xsl:param>
  <xsl:param name="title">TITLE</xsl:param>

  <!-- catch the root element -->
  <xsl:template match="/xhtml">
    <xsl:text disable-output-escaping="yes">
    &lt;!DOCTYPE weblog PUBLIC "-//CAFEBABE//DTD weblog 1.0//EN"
                               "../dtd/weblog.dtd">
    </xsl:text>
    <weblog>
      <xsl:attribute name="id"><xsl:value-of select="$id"/></xsl:attribute>
      <xsl:attribute name="date"><xsl:value-of select="$date"/></xsl:attribute>
      <title><xsl:value-of select="$title"/></title>
      <xsl:apply-templates/>
    </weblog>
  </xsl:template>

  <xsl:template match="h1">
    <p><imp><xsl:apply-templates/></imp></p>
  </xsl:template>

  <xsl:template match="h2">
    <p><imp><xsl:apply-templates/></imp></p>
  </xsl:template>

  <xsl:template match="h3">
    <p><imp><xsl:apply-templates/></imp></p>
  </xsl:template>

  <xsl:template match="h4">
    <p><imp><xsl:apply-templates/></imp></p>
  </xsl:template>

  <xsl:template match="h5">
    <p><imp><xsl:apply-templates/></imp></p>
  </xsl:template>

  <xsl:template match="h6">
    <p><imp><xsl:apply-templates/></imp></p>
  </xsl:template>

  <xsl:template match="p">
    <p><xsl:apply-templates/></p>
  </xsl:template>

  <xsl:template match="ul">
    <list><xsl:apply-templates/></list>
  </xsl:template>

  <xsl:template match="ol">
    <enum><xsl:apply-templates/></enum>
  </xsl:template>

  <xsl:template match="li">
    <item><xsl:apply-templates/></item>
  </xsl:template>

  <xsl:template match="table">
    <table><xsl:apply-templates/></table>
  </xsl:template>

  <xsl:template match="th">
    <th><xsl:apply-templates/></th>
  </xsl:template>

  <xsl:template match="tr">
    <li><xsl:apply-templates/></li>
  </xsl:template>

  <xsl:template match="td">
    <co><xsl:apply-templates/></co>
  </xsl:template>

  <xsl:template match="pre">
    <source><xsl:apply-templates/></source>
  </xsl:template>

  <xsl:template match="img">
    <figure url="{@src}" width="{@width}" height="{@height}"/>
  </xsl:template>

  <xsl:template match="em">
    <term><xsl:apply-templates/></term>
  </xsl:template>

  <xsl:template match="strong">
    <imp><xsl:apply-templates/></imp>
  </xsl:template>

  <xsl:template match="code">
    <code><xsl:apply-templates/></code>
  </xsl:template>

</xsl:stylesheet>`
)

func processXsl(xmlFile string, data map[string]string) ([]byte) {
    xslFile, err := ioutil.TempFile("/tmp", "md2xsl-")
    if err != nil {
        panic(err)
    }
    err = ioutil.WriteFile(xslFile.Name(), []byte(STYLESHEET), 0755)
    if err != nil {
        panic(err)
    }
    defer os.Remove(xslFile.Name())
    params := make([]string, 0, 2 + 3*len(data))
    for name, value := range data {
        params = append(params, "--stringparam")
        params = append(params, name)
        params = append(params, value)
    }
    params = append(params, xslFile.Name())
    params = append(params, xmlFile)
    command := exec.Command("xsltproc", params...)
    result, err := command.CombinedOutput()
    if err != nil {
        panic(err)
    }
    return result
}

func markdown2xhtml(markdown string) ([]byte) {
    xhtml := "<xhtml>\n" + string(blackfriday.MarkdownCommon([]byte(markdown))) + "\n</xhtml>"
    return []byte(xhtml)
}

func markdownData(text string) (map[string]string, string) {
    data := make(map[string]string)
    lines := strings.Split(text, "\n")
    var limit int
    for index, line := range lines {
        if strings.HasPrefix(line, "% ") && strings.Index(line, ":") >= 0 {
            name := strings.TrimSpace(line[2:strings.Index(line, ":")])
            value := strings.TrimSpace(line[strings.Index(line, ":")+1:len(line)])
            data[name] = value
        } else {
            limit = index
            break
        }
    }
    return data, strings.Join(lines[limit:len(lines)], "\n")
}

func processFile(filename string) string {
    source, err := ioutil.ReadFile(filename)
    if err != nil {
        panic(err)
    }
    data, markdown := markdownData(string(source))
    xhtml := markdown2xhtml(markdown)
    xmlFile, err := ioutil.TempFile("/tmp", "md2xml-")
    if err != nil {
        panic(err)
    }
    defer os.Remove(xmlFile.Name())
    ioutil.WriteFile(xmlFile.Name(), xhtml, 0755)
    result := processXsl(xmlFile.Name(), data)
    return string(result)
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println(HELP)
        os.Exit(1)
    } else if os.Args[1] == "-h" || os.Args[1] == "--help" {
        fmt.Println(HELP)
        os.Exit(0)
    } else {
        for _, filename := range os.Args[1:len(os.Args)] {
            fmt.Println(processFile(filename))
        }
    }
}
