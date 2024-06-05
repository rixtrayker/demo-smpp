#!/usr/bin/env fish

function concat_go_files_by_package
    # Check if directory path is provided as an argument
    if test -n "$argv"
        set dir_path $argv[1]
    else
        set dir_path "."
    end

    # Generate output file name
    set base_name (basename $dir_path)
    set output_file "pkg-$base_name.go"

    # Find all *.go files recursively in the specified directory
    find $dir_path -type f -name "*.go" | while read file
        echo "File: $file" >> $output_file
        cat $file >> $output_file
        echo " "
        echo "_____________________" >> $output_file
    end

    echo "Go files concatenated by package!"
end

# Call the function
concat_go_files_by_package $argv
