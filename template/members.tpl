{{ define "members" }}

{{ range .Members }}
{{ if not (hiddenMember .)}}
<tr>
    <td>
        <code>{{ fieldName . }}</code></br>
        <em>
            {{ if linkForType .Type }}
                <a href="{{linkForType .Type}}">
                    {{ typeDisplayName .Type }}
                </a>
            {{ else }}
                {{ typeDisplayName .Type }}
            {{ end }}
        </em>
    </td>
    <td>
        {{ if fieldEmbedded . }}
            <p>
                (Members of <code>{{ fieldName . }}</code> are embedded into this type.)
            </p>
        {{ end}}
    {{ safe (nl2br (showComment .CommentLines)) }}

    {{ if or (eq (fieldName .) "spec") }}
        <br/>
        <br/>
        <table>
            {{ template "members" .Type }}
        </table>
    {{ end }}
    </td>
</tr>
{{ end }}
{{ end }}

{{ end }}
